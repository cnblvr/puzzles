package frontend

import (
	"io"
	"net/http"
)

func (srv *service) HandleLogout(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "TODO logout") // TODO
}
