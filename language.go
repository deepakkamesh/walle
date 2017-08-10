package walle

import (
	"context"

	language "cloud.google.com/go/language/apiv1"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"
)

// Analyze Sentiment analyzes the sentiment of the txt string and returns the sentiment and
// magnitude of the sentiment.
func AnalyzeSentiment(txt string) (score float32, magnitude float32, err error) {

	ctx := context.Background()

	// Creates a client.
	client, err := language.NewClient(ctx)
	if err != nil {
		return
	}

	// Sets the text to analyze.

	// Detects the sentiment of the text.
	req := &languagepb.AnalyzeSentimentRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: txt,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	}

	sentiment, err := client.AnalyzeSentiment(ctx, req)
	if err != nil {
		return
	}

	score = sentiment.DocumentSentiment.Score
	magnitude = sentiment.DocumentSentiment.Magnitude
	return
}
