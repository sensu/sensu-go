package e2e

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func waitForBackend(url string) bool {
	for i := 0; i < 10; i++ {
		resp, getErr := http.Get(fmt.Sprintf("%s/health", url))
		if getErr != nil {
			log.Println("backend not ready, sleeping...")
			time.Sleep(1 * time.Second)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode != 200 && resp.StatusCode != 401 {
			log.Printf("backend returned non-200/401 status code: %d\n", resp.StatusCode)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Println("backend ready")
		return true
	}
	return false
}
