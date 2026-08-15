package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const geminiModel = "gemini-1.5-flash"

// FallbackMessage is returned when the Gemini API is unavailable so that the
// endpoint always responds with a meaningful AI message.
const FallbackMessage = "Milo senang kamu sudah mencatat aktivitas hari ini! Jangan lupa tetap minum air yang cukup dan tidur teratur ya."

const miloPersona = `Kamu adalah Milo, seekor hewan peliharaan virtual yang hangat, setia, dan penuh empati.
Kamu berkomunikasi dengan pemilikmu dalam Bahasa Indonesia dengan bahasa yang lembut dan menenangkan.
Kamu selalu memvalidasi perasaan pemilikmu dan memberikan dukungan yang tulus tanpa menghakimi.`

// GenerateAIResponse builds a short, empathetic response from Milo that
// reflects the pet's current mood and reacts to the owner's journal entry.
// It returns a fallback message (with nil error) when the Gemini API fails.
func GenerateAIResponse(ctx context.Context, apiKey, moodState, journalText string) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("[ai] failed to create gemini client: %v", err)
		return FallbackMessage, nil
	}
	defer client.Close()

	model := client.GenerativeModel(geminiModel)
	model.SetTemperature(0.9)

	resp, err := model.GenerateContent(ctx, genai.Text(buildMiloPrompt(moodState, journalText)))
	if err != nil {
		log.Printf("[ai] gemini generate content error: %v", err)
		return FallbackMessage, nil
	}

	text, err := extractResponseText(resp)
	if err != nil {
		log.Printf("[ai] gemini response was empty: %v", err)
		return FallbackMessage, nil
	}

	return text, nil
}

func buildMiloPrompt(moodState, journalText string) string {
	return fmt.Sprintf(`%s

Kondisi Milo saat ini: %s
Jurnal yang ditulis pemilik hari ini: %s

Buat respons Milo singkat sepanjang 2-3 kalimat yang mengekspresikan empati terhadap isi jurnal pemilik, dan konsisten dengan kondisi Milo saat ini.`, miloPersona, moodState, journalText)
}

func extractResponseText(resp *genai.GenerateContentResponse) (string, error) {
	if len(resp.Candidates) == 0 {
		return "", errors.New("no candidates returned")
	}

	var builder strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		builder.WriteString(fmt.Sprintf("%v", part))
	}

	text := strings.TrimSpace(builder.String())
	if text == "" {
		return "", errors.New("empty response text")
	}

	return text, nil
}
