package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/vishnuvyas/banno/internal"
	"github.com/vishnuvyas/banno/internal/dropbox"
)

func getSignatureList(api dropbox.DropboxAPIClient) {
	const PageSize int = 40
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var signatureRequests []dropbox.SignatureRequest
	resp, err := api.GetSignatureRequests(ctx, "all", 1, PageSize, "")
	if err != nil {
		log.Fatalf("error while getting signatures: %v", err)
	}

	signatureRequests = append(signatureRequests, resp.SignatureRequests...)
	log.Printf("Consumed %d pages out of %d", 1, resp.ListInfo.NumPages)

	for pageNum := range resp.ListInfo.NumPages - 1 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := api.GetSignatureRequests(ctx, "all", pageNum+2, PageSize, "")
		if err != nil {
			log.Fatalf("error while getting signatures on page %d (%v)", pageNum+2, err)
		}
		signatureRequests = append(signatureRequests, resp.SignatureRequests...)
		log.Printf("Consumed %d pages of %d", pageNum+2, resp.ListInfo.NumPages)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	defer w.Flush()
	header := strings.Join([]string{
		"SignatureID",
		"IsComplete",
		"FileURL",
	}, "\t")
	fmt.Fprintln(w, header)

	for _, req := range signatureRequests {
		rowStr := strings.Join([]string{
			req.SignatureRequestID,
			fmt.Sprintf("%t", req.IsComplete),
			req.FilesURL,
		}, "\t")
		fmt.Fprintln(w, rowStr)
	}
}

func main() {
	configFilename := flag.String("c", "config.toml", "config toml file name to use")
	command := flag.String("x", "signature-list", "command to execute")
	flag.Parse()

	cfg, err := internal.NewConfig(*configFilename)
	if err != nil {
		log.Fatalf("unable to load configuration file: %s (%v)", *configFilename, err)
	}

	api, err := dropbox.NewAPIClient(cfg)
	if err != nil {
		log.Fatalf("unable to create api client: %v", err)
	}

	switch *command {
	case "signature-list":
		getSignatureList(api)
	default:
		log.Fatalf("unsupported command: %s", *command)
	}
}
