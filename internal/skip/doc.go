// Package skip implements a streaming filter that silently discards
// the first N chunks written to it and forwards all subsequent chunks
// to the wrapped destination writer.
//
// This is useful when replaying a snapshot and you want to resume
// processing after a known offset — for example, skipping chunks that
// have already been processed by a downstream consumer.
//
// Basic usage:
//
//	s, err := skip.New(dst, 10)
//	if err != nil {
//		log.Fatal(err)
//	}
//	// The first 10 Write calls are dropped; the rest pass through.
//	for _, chunk := range chunks {
//		s.Write(chunk)
//	}
//	fmt.Printf("dropped=%d passed=%d\n", s.Dropped(), s.Passed())
package skip
