package keeper_test

import (
	"context"
	"testing"

	"github.com/alice/checkers/x/checkers"
	"github.com/alice/checkers/x/checkers/keeper"
	"github.com/alice/checkers/x/checkers/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const (
	alice = "cosmos1jmjfq0tplp9tmx4v9uemw72y4d2wa5nr3xn9d3"
	bob   = "cosmos1xyxs3skf3f4jfqeuv89yyaqvjc6lffavxqhc8g"
	carol = "cosmos1e0w5t53nrq7p66fye6c8p0ynyhf6y24l4yuxd7"
)

func TestCreateGame(t *testing.T) {
	msgServer, _, context := setupMsgServerCreateGame(t)
	createResponse, err := msgServer.CreateGame(context, &types.MsgCreateGame{
		Creator: alice,
		Red:     bob,
		Black:   carol,
	})
	require.Nil(t, err)
	require.EqualValues(t, types.MsgCreateGameResponse{
		IdValue: "1",
	}, *createResponse)
}

func setupMsgServerCreateGame(t testing.TB) (types.MsgServer, keeper.Keeper, context.Context) {
	k, ctx := setupKeeper(t)
	checkers.InitGenesis(ctx, *k, *types.DefaultGenesis())
	return keeper.NewMsgServerImpl(*k), *k, sdk.WrapSDKContext(ctx)
}

func TestCreate1GameHasSaved(t *testing.T) {
	msgSrvr, keeper, context := setupMsgServerCreateGame(t)
	msgSrvr.CreateGame(context, &types.MsgCreateGame{
		Creator: alice,
		Red:     bob,
		Black:   carol,
	})
	nextGame, found := keeper.GetNextGame(sdk.UnwrapSDKContext(context))
	require.True(t, found)
	require.EqualValues(t, types.NextGame{
		Creator: "",
		IdValue: 2,
		FifoHead: "1",
		FifoTail: "1",
	}, nextGame)
	game1, found1 := keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "1")
	require.True(t, found1)
	require.EqualValues(t, types.StoredGame{
		Creator: alice,
		Index:   "1",
		Game:    "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
		Turn:    "b",
		Red:     bob,
		Black:   carol,
		MoveCount: uint64(0),
		BeforeId: "-1",
		AfterId: "-1",
	}, game1)
}

func TestCreate1GameEmitted(t *testing.T) {
	msgSrvr, _, context := setupMsgServerCreateGame(t)
	msgSrvr.CreateGame(context, &types.MsgCreateGame{
		Creator: alice,
		Red:     bob,
		Black:   carol,
	})
	ctx := sdk.UnwrapSDKContext(context)
	require.NotNil(t, ctx)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 1)
	event := events[0]
	require.EqualValues(t, sdk.StringEvent{
		Type: "message",
		Attributes: []sdk.Attribute{
			{Key: "module", Value: "checkers"},
			{Key: "action", Value: "NewGameCreated"},
			{Key: "Creator", Value: alice},
			{Key: "Index", Value: "1"},
			{Key: "Red", Value: bob},
			{Key: "Black", Value: carol},
		},
	}, event)
}

func TestPlayMoveEmitted(t *testing.T) {
	msgServer, _, context := setupMsgServerWithOneGameForPlayMove(t)
	msgServer.PlayMove(context, &types.MsgPlayMove{
		Creator: carol,
		IdValue: "1",
		FromX:   1,
		FromY:   2,
		ToX:     2,
		ToY:     3,
	})
	ctx := sdk.UnwrapSDKContext(context)
	require.NotNil(t, ctx)
	events := sdk.StringifyEvents(ctx.EventManager().ABCIEvents())
	require.Len(t, events, 1)
	event := events[0]
	require.Equal(t, event.Type, "message")
	require.EqualValues(t, []sdk.Attribute{
		{Key: "module", Value: "checkers"},
		{Key: "action", Value: "MovePlayed"},
		{Key: "Creator", Value: carol},
		{Key: "IdValue", Value: "1"},
		{Key: "CapturedX", Value: "-1"},
		{Key: "CapturedY", Value: "-1"},
		{Key: "Winner", Value: "NO_PLAYER"},
	}, event.Attributes[6:])
}

func TestCreate3GamesHasSavedFifo(t *testing.T) {
	msgSrvr, keeper, context := setupMsgServerCreateGame(t)
	msgSrvr.CreateGame(context, &types.MsgCreateGame{
			Creator: alice,
			Red:     bob,
			Black:   carol,
	})

	msgSrvr.CreateGame(context, &types.MsgCreateGame{
			Creator: bob,
			Red:     carol,
			Black:   alice,
	})
	nextGame2, found2 := keeper.GetNextGame(sdk.UnwrapSDKContext(context))
	require.True(t, found2)
	require.EqualValues(t, types.NextGame{
			Creator:  "",
			IdValue:  3,
			FifoHead: "1",
			FifoTail: "2",
	}, nextGame2)
	game1, found1 := keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "1")
	require.True(t, found1)
	require.EqualValues(t, types.StoredGame{
			Creator:   alice,
			Index:     "1",
			Game:      "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
			Turn:      "b",
			Red:       bob,
			Black:     carol,
			MoveCount: uint64(0),
			BeforeId:  "-1",
			AfterId:   "2",
	}, game1)
	game2, found2 := keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "2")
	require.True(t, found2)
	require.EqualValues(t, types.StoredGame{
			Creator:   bob,
			Index:     "2",
			Game:      "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
			Turn:      "b",
			Red:       carol,
			Black:     alice,
			MoveCount: uint64(0),
			BeforeId:  "1",
			AfterId:   "-1",
	}, game2)

	msgSrvr.CreateGame(context, &types.MsgCreateGame{
			Creator: carol,
			Red:     alice,
			Black:   bob,
	})
	nextGame3, found3 := keeper.GetNextGame(sdk.UnwrapSDKContext(context))
	require.True(t, found3)
	require.EqualValues(t, types.NextGame{
			Creator:  "",
			IdValue:  4,
			FifoHead: "1",
			FifoTail: "3",
	}, nextGame3)
	game1, found1 = keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "1")
	require.True(t, found1)
	require.EqualValues(t, types.StoredGame{
			Creator:   alice,
			Index:     "1",
			Game:      "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
			Turn:      "b",
			Red:       bob,
			Black:     carol,
			MoveCount: uint64(0),
			BeforeId:  "-1",
			AfterId:   "2",
	}, game1)
	game2, found2 = keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "2")
	require.True(t, found2)
	require.EqualValues(t, types.StoredGame{
			Creator:   bob,
			Index:     "2",
			Game:      "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
			Turn:      "b",
			Red:       carol,
			Black:     alice,
			MoveCount: uint64(0),
			BeforeId:  "1",
			AfterId:   "3",
	}, game2)
	game3, found3 := keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "3")
	require.True(t, found3)
	require.EqualValues(t, types.StoredGame{
			Creator:   carol,
			Index:     "3",
			Game:      "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
			Turn:      "b",
			Red:       alice,
			Black:     bob,
			MoveCount: uint64(0),
			BeforeId:  "2",
			AfterId:   "-1",
	}, game3)
}

func TestPlayMove2Games2MovesHasSavedFifo(t *testing.T) {
	msgServer, keeper, context := setupMsgServerWithOneGameForPlayMove(t)
	msgServer.CreateGame(context, &types.MsgCreateGame{
			Creator: bob,
			Red:     carol,
			Black:   alice,
	})
	msgServer.PlayMove(context, &types.MsgPlayMove{
			Creator: carol,
			IdValue: "1",
			FromX:   1,
			FromY:   2,
			ToX:     2,
			ToY:     3,
	})

	msgServer.PlayMove(context, &types.MsgPlayMove{
			Creator: alice,
			IdValue: "2",
			FromX:   1,
			FromY:   2,
			ToX:     2,
			ToY:     3,
	})
	nextGame1, found1 := keeper.GetNextGame(sdk.UnwrapSDKContext(context))
	require.True(t, found1)
	require.EqualValues(t, types.NextGame{
			Creator:  "",
			IdValue:  3,
			FifoHead: "1",
			FifoTail: "2",
	}, nextGame1)
	game1, found1 := keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "1")
	require.True(t, found1)
	require.EqualValues(t, types.StoredGame{
			Creator:   alice,
			Index:     "1",
			Game:      "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
			Turn:      "r",
			Red:       bob,
			Black:     carol,
			MoveCount: uint64(1),
			BeforeId:  "-1",
			AfterId:   "2",
	}, game1)
	game2, found2 := keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "2")
	require.True(t, found2)
	require.EqualValues(t, types.StoredGame{
			Creator:   bob,
			Index:     "2",
			Game:      "*b*b*b*b|b*b*b*b*|***b*b*b|**b*****|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
			Turn:      "r",
			Red:       carol,
			Black:     alice,
			MoveCount: uint64(1),
			BeforeId:  "1",
			AfterId:   "-1",
	}, game2)
}

func TestRejectMiddleGameHasSavedFifo(t *testing.T) {
	msgServer, keeper, context := setupMsgServerWithOneGameForRejectGame(t)
	msgServer.CreateGame(context, &types.MsgCreateGame{
			Creator: bob,
			Red:     carol,
			Black:   alice,
	})
	msgServer.CreateGame(context, &types.MsgCreateGame{
			Creator: carol,
			Red:     alice,
			Black:   bob,
	})
	msgServer.RejectGame(context, &types.MsgRejectGame{
			Creator: carol,
			IdValue: "2",
	})
	nextGame, found := keeper.GetNextGame(sdk.UnwrapSDKContext(context))
	require.True(t, found)
	require.EqualValues(t, types.NextGame{
			Creator:  "",
			IdValue:  4,
			FifoHead: "1",
			FifoTail: "3",
	}, nextGame)
	game1, found1 := keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "1")
	require.True(t, found1)
	require.EqualValues(t, types.StoredGame{
			Creator:   alice,
			Index:     "1",
			Game:      "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
			Turn:      "b",
			Red:       bob,
			Black:     carol,
			MoveCount: uint64(0),
			BeforeId:  "-1",
			AfterId:   "3",
	}, game1)
	game3, found3 := keeper.GetStoredGame(sdk.UnwrapSDKContext(context), "3")
	require.True(t, found3)
	require.EqualValues(t, types.StoredGame{
			Creator:   carol,
			Index:     "3",
			Game:      "*b*b*b*b|b*b*b*b*|*b*b*b*b|********|********|r*r*r*r*|*r*r*r*r|r*r*r*r*",
			Turn:      "b",
			Red:       alice,
			Black:     bob,
			MoveCount: uint64(0),
			BeforeId:  "1",
			AfterId:   "-1",
	}, game3)
}