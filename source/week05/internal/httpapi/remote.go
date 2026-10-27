package httpapi

import (
	"errors"
	"example.edu/6cc545/cw1/internal/config"
	"example.edu/6cc545/cw1/internal/model"
	"example.edu/6cc545/cw1/internal/rpcfiles"
	"example.edu/6cc545/cw1/internal/state"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
)

type remoteFilesResponse struct {
	NodeID string               `json:"node_id"`
	Files  []model.FileMetadata `json:"files"`
}

// mapRPCError turns a gRPC status into an HTTP status and a JSON error label.
func mapRPCError(err error) (int, string) {
	switch status.Code(err) {
	case codes.InvalidArgument:
		return http.StatusBadRequest, "remote_rejected_request"
	case codes.Unavailable:
		return http.StatusServiceUnavailable, "peer_unavailable"
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout, "peer_timeout"
	case codes.Canceled:
		return http.StatusServiceUnavailable, "request_cancelled"
	default:
		return http.StatusBadGateway, "remote_error"
	}
}

func RemoteFiles(reg *state.Registry, nodeID string, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			ErrorJSON(w, http.StatusMethodNotAllowed, "method_not_allowed")
			return
		}
		peerID := r.URL.Query().Get("peer_id")
		if !config.ValidID(peerID) { // missing or malformed
			ErrorJSON(w, http.StatusBadRequest, "invalid_peer_id")
			return
		}
		peer, err := reg.Get(peerID)
		if err != nil {
			if errors.Is(err, state.ErrMissing) {
				ErrorJSON(w, http.StatusNotFound, "peer_not_found")
				return
			}
			ErrorJSON(w, http.StatusInternalServerError, "registry_error")
			return
		}
		// r.Context(): if the HTTP client goes away, the RPC is cancelled too.
		resp, err := rpcfiles.Fetch(r.Context(), peer.GRPCAddress, nodeID, timeout)
		if err != nil {
			code, label := mapRPCError(err)
			ErrorJSON(w, code, label)
			return
		}
		files := make([]model.FileMetadata, 0, len(resp.GetFiles()))
		for _, f := range resp.GetFiles() {
			files = append(files, model.FileMetadata{
				Name:         f.GetName(),
				Size:         f.GetSize(),
				SHA256:       f.GetSha256(),
				ModifiedUnix: f.GetModifiedUnix(),
			})
		}
		WriteJSON(w, http.StatusOK, remoteFilesResponse{NodeID: resp.GetNodeId(), Files: files})
	}
}