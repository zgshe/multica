package handler

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

// HandleAdminUpgrade executes the multica-upgrade.sh script
// and returns the result to the client.
func (h *Handler) HandleAdminUpgrade(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Find the multica scripts directory
	// The backend runs from server/ directory, so scripts are at ../scripts/
	multicaDir := os.Getenv("MULTICA_DIR")
	if multicaDir == "" {
		// Fallback: assume backend runs from server/ directory
		multicaDir = ".."
	}
	scriptPath := multicaDir + "/scripts/multica-upgrade.sh"

	// Execute the upgrade script
	cmd := exec.CommandContext(ctx, "bash", scriptPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("upgrade failed: %v\n%s\n", err, output)
		http.Error(w, fmt.Sprintf("upgrade failed: %v\n%s", err, output), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("upgrade completed:\n%s", output)))
}