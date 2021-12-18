package restapi

import "encoding/base64"

func DecodeCursor(enc string) (dec string) {
	if enc == "" {
		return
	}
	bt, _ := base64.StdEncoding.DecodeString(enc)
	return string(bt)
}

func EncodeCursor(raw string) string {
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

type PaginationResponse struct {
	Cursor   string `json:"cursor"`
	Backward bool   `json:"backward"`
	HasNext  bool   `json:"hasNext"`
	Count    uint64 `json:"count"`
	Size     uint64 `json:"size"`
}

func NewPaginationResponse(cursor string, backward bool, count uint64, size uint64, lenData int) PaginationResponse {
	return PaginationResponse{
		Cursor:   cursor,
		Backward: backward,
		Count:    count,
		HasNext:  lenData > 0,
		Size:     size,
	}
}

type PaginationRequest struct {
	Cursor   string `json:"cursor"`
	Backward bool   `json:"backward"`
	Size     uint64 `json:"size"`
}

type ResponseWithPagination struct {
	Pagination PaginationResponse `json:"pagination,omitempty"`
	// Will contains array of any object
	Data interface{} `json:"data" swaggertype:"array,object"`
}
