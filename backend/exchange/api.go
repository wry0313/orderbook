package exchange

import (
	"encoding/json"
	"github/wry-0313/exchange/endpoint"
	ws "github/wry-0313/exchange/websocket"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

const (
	ErrMsgInternalServer = "Internal server error"
)


type API struct {
	exchangeService Service
}

func NewAPI(exchangeService Service) *API {
	return &API{
		exchangeService: exchangeService,
	}
}

func (api *API) HandlePlaceOrder(w http.ResponseWriter, r *http.Request) {
	var order PlaceOrderInput
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		log.Printf("handler: failed to decode request: %v\n", err)
		endpoint.HandleDecodeErr(w, err)
		return
	}

	defer r.Body.Close()

	err := api.exchangeService.PlaceOrder(order)
	if err != nil {
		endpoint.WriteWithError(w, http.StatusInternalServerError, ErrMsgInternalServer)
		return
	}
}

func (api *API) HandleStreamMarketPrice(w http.ResponseWriter, r *http.Request) {
	var input StreamPriceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("handler: failed to decode request: %v\n", err)
		endpoint.HandleDecodeErr(w, err)
		return
	}

	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p, err := api.exchangeService.GetMarketPrice(input.Symbol)
			if err != nil {
				log.Printf("handler: failed to get market price: %v\n", err)
				endpoint.WriteWithError(w, http.StatusInternalServerError, ErrMsgInternalServer)
				return
			}
			priceString := p.String()

			if err := conn.WriteMessage(websocket.TextMessage, []byte(priceString)); err != nil {
				log.Println("WriteMessage Error:", err)
				return
			}
		}
	}
}

// RegisterHandlers is a function that registers all the handlers for the user endpoints
func (api *API) RegisterHandlers(r chi.Router, authHandler func(http.Handler) http.Handler) {
	r.Route("/order", func(r chi.Router) {
		r.Post("/", api.HandlePlaceOrder)
	})
}
