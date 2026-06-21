package http

import stdhttp "net/http"

// HTTPステータスコード定数
const (
	StatusNoContent           = stdhttp.StatusNoContent           // 204
	StatusBadRequest          = stdhttp.StatusBadRequest          // 400
	StatusConflict            = stdhttp.StatusConflict            // 409
	StatusInternalServerError = stdhttp.StatusInternalServerError // 500
	StatusBadGateway          = stdhttp.StatusBadGateway          // 502
)
