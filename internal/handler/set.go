package handler

import (
	"github.com/goforj/wire"

	handlerhttp "github.com/opsybot/opsybot/internal/handler/http"
	"github.com/opsybot/opsybot/internal/handler/http/v1/dashboard"
)

var Set = wire.NewSet(
	dashboard.New,
	handlerhttp.NewRouter,
)
