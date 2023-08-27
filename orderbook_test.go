package orderbook

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestPlaceMarketOrderAfterLimit(t *testing.T) {
	ob := NewOrderBook()
	clientID := uuid.New()
	for i := 0; i < 2; i++ {
		ob.PlaceLimitOrder(Buy, clientID, decimal.NewFromInt(10000), decimal.NewFromInt(10))
	}
	max, _ := ob.bids.priceTree.GetMax()
	fmt.Printf("max: %v\n", max)
	ob.PlaceMarketOrder(Sell, clientID, decimal.NewFromInt(50))
	ob.PlaceMarketOrder(Sell, clientID, decimal.NewFromInt(50))
	ob.PlaceMarketOrder(Sell, clientID, decimal.NewFromInt(50))
	ob.PlaceMarketOrder(Sell, clientID, decimal.NewFromInt(50))
	ob.PlaceMarketOrder(Sell, clientID, decimal.NewFromInt(50))
	ob.PlaceMarketOrder(Sell, clientID, decimal.NewFromInt(9751))
	fmt.Printf("max: %v\n", max)
}

func TestPlaceMarketOrderAfterLimitConcurrent(t *testing.T) {
	ob := NewOrderBook()
	clientID := uuid.New()
	// ch := make(chan int)
	var wg sync.WaitGroup
	orderID, err := ob.PlaceLimitOrder(Buy, clientID, decimal.NewFromInt(10000), decimal.NewFromInt(10))
	if err != nil {
		t.Error(err)
	}
	max, _ := ob.bids.priceTree.GetMax()
	fmt.Printf("max: %v\n", max)
	time.Sleep(2 * time.Second)
	wg.Add(1)
	go func() {
		fmt.Println("start1")
		defer wg.Done()
		for i := 0; i < 10; i++ {
			ob.PlaceLimitOrder(Buy, clientID, decimal.NewFromInt(10), decimal.NewFromInt(10))
		}
	}()
	wg.Add(1)
	go func() {
		fmt.Println("start1.5")
		defer wg.Done()
		for i := 0; i < 5; i++ {
			ob.PlaceLimitOrder(Buy, clientID, decimal.NewFromInt(10), decimal.NewFromInt(10))
		}
	}()
	wg.Add(1)
	go func() {
		fmt.Println("start2")
		defer wg.Done()
		for i := 0; i < 51; i++ {
			ob.PlaceMarketOrder(Sell, clientID, decimal.NewFromInt(1))
		}
	}()
	wg.Wait()
	fmt.Printf("max: %v\n", max)
	fmt.Printf("order: %v\n", ob.activeOrders[orderID].Value.logs)
	fmt.Printf("orderside volume %v\n", ob.bids.volume)
}

func TestMarketOrderPartialFill(t *testing.T) {
	ob := NewOrderBook()
	clientID := uuid.New()
	_, err := ob.PlaceLimitOrder(Buy, clientID, decimal.NewFromInt(10), decimal.NewFromInt(10))
	if err != nil {
		t.Error(err)
	}
	ob.PlaceMarketOrder(Sell, clientID, decimal.NewFromInt(15))
	// Log(fmt.Sprintf("order side volume %v\n", ob.asks.volume))
	// Log(fmt.Sprintf("order queue volume %v\n", oq.volume))
	assert(t, ob.asks.volume.String(), "5")
	assert(t, ob.bids.volume.String(), "0")
	depth := ob.asks.depth
	assert(t, depth, 0)
}

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}
