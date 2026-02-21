package rag

import (
	"log"
	"strings"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

var defaultSeparators = []string{"\n\n", "\n", ". ", " "}

// tokenEncoder is initialized once for cl100k_base (used by text-embedding-3-small).
var tokenEncoder *tiktoken.Tiktoken

func init() {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		log.Fatalf("failed to load tiktoken encoding: %v", err)
	}
	tokenEncoder = enc
}

// tokenLen returns the number of tokens in the text using cl100k_base encoding.
func tokenLen(text string) int {
	return len(tokenEncoder.Encode(text, nil, nil))
}

func ChunkText(text string, chunkSize, overlap int) []string {
	if chunkSize <= 0 {
		chunkSize = 500
	}
	if overlap <= 0 {
		overlap = 100
	}

	return recursiveSplit(text, chunkSize, overlap, defaultSeparators)
}

func recursiveSplit(text string, chunkSize, overlap int, separators []string) []string {
	if tokenLen(text) <= chunkSize {
		if strings.TrimSpace(text) == "" {
			return nil
		}
		return []string{text}
	}

	// No separators left — fall back to token-based splitting
	if len(separators) == 0 {
		return splitByTokens(text, chunkSize, overlap)
	}

	sep := separators[0]
	parts := strings.Split(text, sep)

	var chunks []string
	var current strings.Builder

	for i, part := range parts {
		piece := part
		if i > 0 {
			piece = sep + part
		}

		if tokenLen(current.String()+piece) > chunkSize && current.Len() > 0 {
			chunk := current.String()
			// If this merged chunk is still too big, recurse with finer separators
			if tokenLen(chunk) > chunkSize {
				chunks = append(chunks, recursiveSplit(chunk, chunkSize, overlap, separators[1:])...)
			} else {
				chunks = append(chunks, chunk)
			}

			// Start new chunk with overlap from end of previous
			current.Reset()
			overlapText := getOverlapSuffix(chunk, overlap)
			current.WriteString(overlapText)
		}
		current.WriteString(piece)
	}

	// Flush remaining
	if current.Len() > 0 {
		remaining := current.String()
		if strings.TrimSpace(remaining) != "" {
			if tokenLen(remaining) > chunkSize {
				chunks = append(chunks, recursiveSplit(remaining, chunkSize, overlap, separators[1:])...)
			} else {
				chunks = append(chunks, remaining)
			}
		}
	}

	return chunks
}

func splitByTokens(text string, chunkSize, overlap int) []string {
	tokens := tokenEncoder.Encode(text, nil, nil)
	length := len(tokens)
	if length == 0 {
		return nil
	}

	var chunks []string
	start := 0
	for start < length {
		end := min(start+chunkSize, length)
		chunk := tokenEncoder.Decode(tokens[start:end])
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, chunk)
		}
		start += chunkSize - overlap
	}
	return chunks
}

func getOverlapSuffix(text string, overlap int) string {
	tokens := tokenEncoder.Encode(text, nil, nil)
	if len(tokens) <= overlap {
		return text
	}
	return tokenEncoder.Decode(tokens[len(tokens)-overlap:])
}
