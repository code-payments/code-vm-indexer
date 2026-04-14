// indexer-cli is a small utility for hitting the Indexer gRPC service.
//
// Examples:
//
//	indexer-cli -addr localhost:8086 timelock -vm <base58> -owner <base58>
//	indexer-cli -addr localhost:8086 nonce    -vm <base58> -address <base58>
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mr-tron/base58"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	indexerpb "github.com/code-payments/code-vm-indexer/generated/indexer/v1"
)

func main() {
	rootFlags := flag.NewFlagSet("indexer-cli", flag.ExitOnError)
	addr := rootFlags.String("addr", "localhost:8086", "server address (host:port)")
	timeout := rootFlags.Duration("timeout", 5*time.Second, "request timeout")
	useTLS := rootFlags.Bool("tls", false, "use TLS (system root CAs)")
	rootFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: indexer-cli [global flags] <timelock|nonce> [subcommand flags]\n\n")
		fmt.Fprintln(os.Stderr, "global flags:")
		rootFlags.PrintDefaults()
	}

	if err := rootFlags.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if rootFlags.NArg() < 1 {
		rootFlags.Usage()
		os.Exit(2)
	}

	sub := rootFlags.Arg(0)
	subArgs := rootFlags.Args()[1:]

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	var creds credentials.TransportCredentials
	if *useTLS {
		creds = credentials.NewTLS(&tls.Config{})
	} else {
		creds = insecure.NewCredentials()
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := indexerpb.NewIndexerClient(conn)

	switch sub {
	case "timelock":
		runTimelock(ctx, client, subArgs)
	case "nonce":
		runNonce(ctx, client, subArgs)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n", sub)
		rootFlags.Usage()
		os.Exit(2)
	}
}

func runTimelock(ctx context.Context, client indexerpb.IndexerClient, args []string) {
	fs := flag.NewFlagSet("timelock", flag.ExitOnError)
	vm := fs.String("vm", "", "VM account (base58)")
	owner := fs.String("owner", "", "owner address (base58)")
	_ = fs.Parse(args)

	if *vm == "" || *owner == "" {
		fs.Usage()
		os.Exit(2)
	}

	resp, err := client.GetVirtualTimelockAccounts(ctx, &indexerpb.GetVirtualTimelockAccountsRequest{
		VmAccount: mustAddress("vm", *vm),
		Owner:     mustAddress("owner", *owner),
	})
	if err != nil {
		fatalf("rpc: %v", err)
	}
	printJSON(resp)
}

func runNonce(ctx context.Context, client indexerpb.IndexerClient, args []string) {
	fs := flag.NewFlagSet("nonce", flag.ExitOnError)
	vm := fs.String("vm", "", "VM account (base58)")
	address := fs.String("address", "", "nonce address (base58)")
	_ = fs.Parse(args)

	if *vm == "" || *address == "" {
		fs.Usage()
		os.Exit(2)
	}

	resp, err := client.GetVirtualDurableNonce(ctx, &indexerpb.GetVirtualDurableNonceRequest{
		VmAccount: mustAddress("vm", *vm),
		Address:   mustAddress("address", *address),
	})
	if err != nil {
		fatalf("rpc: %v", err)
	}
	printJSON(resp)
}

func mustAddress(name, s string) *indexerpb.Address {
	raw, err := base58.Decode(s)
	if err != nil {
		fatalf("decode %s: %v", name, err)
	}
	return &indexerpb.Address{Value: raw}
}

func printJSON(m proto.Message) {
	out, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(m)
	if err != nil {
		fatalf("marshal: %v", err)
	}
	fmt.Println(string(out))
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
