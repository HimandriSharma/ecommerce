package cart

import (
	"fmt"
	"github.com/HimandriSharma/ecommerce/service/auth"
	"github.com/HimandriSharma/ecommerce/types"
	"github.com/HimandriSharma/ecommerce/utils"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"net/http"
)

type Handler struct {
	store        types.OrderStore
	productStore types.ProductStore
	userStore    types.UserStore
}

func NewHandler(store types.OrderStore, productStore types.ProductStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, productStore: productStore, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/cart/checkout", auth.WithJWTAuth(h.handleCheckout, h.userStore)).Methods(http.MethodPost)
}

func (h *Handler) handleCheckout(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	var cart types.CartCheckoutPayload
	if err := utils.ParseJSON(r, &cart); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(cart); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload for %v", errors))
		return
	}

	productIDs, err := getCartItemIDs(cart.Items)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	ps, err := h.productStore.GetProductByIDs(productIDs)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	orderID, totalPrice, err := h.createOrder(ps, cart.Items, userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"total_price": totalPrice,
		"order_id":    orderID,
	})
}
