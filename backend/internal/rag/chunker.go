package rag

func ChunkText(text string, chunkSize, overlap int) []string {
	if chunkSize <= 0 {
		chunkSize = 800
	}
	if overlap <= 0 {
		overlap = 100
	}

	runes := []rune(text)
	length := len(runes)
	if length == 0 {
		return nil
	}

	var chunks []string
	start := 0

	for start < length {
		end := start + chunkSize
		if end > length {
			end = length
		}
		chunk := string(runes[start:end])
		chunks = append(chunks, chunk)
		start += chunkSize - overlap
	}

	return chunks
}
