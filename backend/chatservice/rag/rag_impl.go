
type defaultPipeline struct {
	ex Extractor
	ch Chunker
	em Embedder
}

func NewPipeline(ex Extractor, ch Chunker, em Embedder) Pipeline {
	return &defaultPipeline{ex: ex, ch: ch, em: em}
}

func (p *defaultPipeline) Run(ctx context.Context, r io.Reader, mime string) ([]Embedding, error) {
	docs, err := p.ex.Extract(ctx, r, mime)
	if err != nil {
		return nil, err
	}

	chunks, err := p.ch.Chunk(ctx, docs)
	if err != nil {
		return nil, err
	}

	return p.em.Embed(ctx, chunks)
}

// ------

// -- Future uses --
// Apache Tika is a great java based library for extracting text from any file, it can be hosted as a server
// TikaExtractor uses Apache Tika-server for any MIME
type TikaExtractor struct{ Endpoint string }

type TextExtractor struct{}
func (e *TextExtractor) Extract(ctx context.Context, r io.Reader, mime string) (Document, error) {
	return Document{
		ID:       "text",
		MIME:     mime,
		Text:     "text",
		Metadata: map[string]string{},
	}

// EqualSizeChunker splits the text into chunks of equal size
type EqualSizeChunker struct{ ChunkSize int }

// FixedParagraphChunker splits on double newlines with some cap
type FixedParagraphChunker struct{ TokenLimit int }

// OLLamaEmbedder hits /v1/embeddings with batching
type OLLamaEmbedder struct{ Model, APIKey string }



// Sample Usage
pipeline := NewPipeline(
	&TextExtractor{},
	&EqualSizeChunker{ChunkSize: 512},
	&OLLamaEmbedder{Model: "bge-base-en", APIKey: "ollama"},
)

pipeline.Run(ctx, r, mime)




