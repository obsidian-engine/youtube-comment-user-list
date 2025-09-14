package http

import stdhttp "net/http"

// HTTPステータスコード定数
const (
	StatusOK                  = stdhttp.StatusOK                  // 200
	StatusNoContent           = stdhttp.StatusNoContent           // 204
	StatusBadRequest          = stdhttp.StatusBadRequest          // 400
	StatusInternalServerError = stdhttp.StatusInternalServerError // 500
	StatusBadGateway          = stdhttp.StatusBadGateway          // 502
)