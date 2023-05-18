package router

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cosmlib "pkg.berachain.dev/polaris/cosmos/lib"
	testutil "pkg.berachain.dev/polaris/cosmos/testing/utils"
	"pkg.berachain.dev/polaris/lib/utils"

	"github.com/argus-labs/world-engine/chain/precompile"
	"github.com/argus-labs/world-engine/chain/router"
	"github.com/argus-labs/world-engine/chain/router/mocks"
	bindings "pkg.berachain.dev/polaris/contracts/bindings/cosmos/precompile"
)

func TestRouterPrecompile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cosmos/precompile/router")
}

var _ = Describe("Router precompile", func() {
	var (
		ctx      sdk.Context
		caller   sdk.AccAddress
		mockCtrl *gomock.Controller
		contract *Contract
		rtr      *mocks.MockRouter
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		rtr = mocks.NewMockRouter(mockCtrl)
		ctx = testutil.NewContext()
		caller = sdk.AccAddress("bob")
		contract = utils.MustGetAs[*Contract](NewPrecompileContract(rtr))
	})

	When("Sending a message", func() {
		It("should fail if there are not enough arguments", func() {
			res, err := contract.Send(
				ctx,
				nil,
				cosmlib.AccAddressToEthAddress(caller),
				big.NewInt(0),
				false,
				"invalid",
			)
			Expect(err).To(MatchError(precompile.ErrInvalidArgumentAmount(2, 1)))
			Expect(res).To(BeNil())
		})
		It("should fail if the first arg is the wrong type", func() {
			res, err := contract.Send(
				ctx,
				nil,
				cosmlib.AccAddressToEthAddress(caller),
				big.NewInt(0),
				false,
				"foo", "bar",
			)
			Expect(err).To(MatchError(precompile.ErrInvalidArgType("[]byte", "foo", 0)))
			Expect(res).To(BeNil())
		})
		It("should fail if the second arg is the wrong type", func() {

			res, err := contract.Send(
				ctx,
				nil,
				cosmlib.AccAddressToEthAddress(caller),
				big.NewInt(0),
				false,
				[]byte("foo"), 15,
			)
			Expect(err).To(MatchError(precompile.ErrInvalidArgType("string", []uint8{}, 1)))
			Expect(res).To(BeNil())
		})
		It("should succeed", func() {
			msg := []byte("foo")
			namespace := "cardinal"
			sender := cosmlib.AccAddressToEthAddress(caller)
			result := router.Result{
				Code:    0,
				Message: []byte("foobar"),
			}
			rtr.EXPECT().Send(ctx, namespace, sender.String(), msg).Times(1).Return(result, nil)
			res, err := contract.Send(
				ctx,
				nil,
				cosmlib.AccAddressToEthAddress(caller),
				big.NewInt(0),
				false,
				msg, namespace,
			)
			Expect(err).ToNot(HaveOccurred())
			gotResult, _ := utils.GetAs[bindings.IRouterResponse](res[0])
			Expect(gotResult).To(Equal(bindings.IRouterResponse{
				Code:    big.NewInt(int64(result.Code)),
				Message: result.Message,
			}))
		})
	})
})