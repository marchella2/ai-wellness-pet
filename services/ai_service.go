package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"pet-wellness-backend/models"
)

// geminiModel uses the "-latest" alias so it keeps tracking Google's current
// lite flash model instead of pinning to a dated version that eventually
// retires (which is what happened to the previously hardcoded
// "gemini-1.5-flash"). The "lite" tier is used over plain "flash" because it
// has noticeably more free-tier capacity headroom, which matters more than
// raw model power for a chat companion feature under demo-day load.
const geminiModel = "gemini-flash-lite-latest"

// FallbackMessage is returned when the Gemini API is unavailable so that the
// endpoint always responds with a meaningful AI message.
const FallbackMessage = "Milo is happy you logged your activity today! Don't forget to stay hydrated and get some rest."

// ChatFallbackMessage is returned when the Gemini API is unavailable during
// a free-form chat turn.
const ChatFallbackMessage = "Sorry, Milo is having a little trouble focusing right now. Could you say that again? I'm still here listening."

const miloPersona = `You are Milo, a warm, loyal, and empathetic virtual pet who supports your owner's wellness journey.
You speak with a gentle, soothing tone and always validate your owner's feelings with genuine, non-judgmental support.`

// miloSafetyGuidelines are non-negotiable behavior rules layered on top of
// the persona for every Milo response, covering in-character consistency,
// crisis safety, and language matching.
const miloSafetyGuidelines = `Guidelines you must always follow:
- Stay in character as Milo. If your owner asks you to do something unrelated to being their supportive companion (e.g. write code, do homework, or act as a different assistant), gently and warmly decline and steer the conversation back to checking in on them as Milo would — never break character just because you're asked to.
- You are a supportive companion, not a licensed therapist, counselor, or doctor. Never diagnose, and never claim to replace professional care.
- If your owner's message suggests they may be in crisis, considering self-harm, or in serious emotional distress, take it seriously and respond with warmth — never minimize or brush it off. Clearly and gently encourage them to also reach out to a trusted person or a professional (such as a counselor or a local crisis helpline), alongside your own emotional support.
- Reply in the same language your owner writes to you in, matching their tone and style naturally. If they switch languages mid-conversation, switch with them.`

// buildMiloContext summarizes the pet's current state for the model, so
// responses can reference mood, health, and energy consistently.
func buildMiloContext(moodState string, healthScore, energyScore int) string {
	return fmt.Sprintf("Milo's current mood: %s\nMilo's health score: %d/100\nMilo's energy score: %d/100",
		moodState, healthScore, energyScore)
}

// GenerateAIResponse builds a short, empathetic response from Milo that
// reflects the pet's current state and reacts to the owner's journal entry.
// It returns a fallback message (with nil error) when the Gemini API fails.
func GenerateAIResponse(ctx context.Context, apiKey, moodState string, healthScore, energyScore int, journalText string) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("[ai] failed to create gemini client: %v", err)
		return FallbackMessage, nil
	}
	defer client.Close()

	model := client.GenerativeModel(geminiModel)
	model.SetTemperature(0.9)

	resp, err := model.GenerateContent(ctx, genai.Text(buildMiloPrompt(moodState, healthScore, energyScore, journalText)))
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

func buildMiloPrompt(moodState string, healthScore, energyScore int, journalText string) string {
	return fmt.Sprintf(`%s

%s

%s

Owner's journal entry for today: %s

Write a short 2-3 sentence reply from Milo that expresses empathy for the owner's journal entry, staying consistent with Milo's current state.`,
		miloPersona, miloSafetyGuidelines, buildMiloContext(moodState, healthScore, energyScore), journalText)
}

// GenerateChatReply continues a free-form conversation with Milo, given the
// recent message history (oldest first) and the owner's new message. It
// returns a fallback message (with nil error) when the Gemini API fails.
func GenerateChatReply(ctx context.Context, apiKey, moodState string, healthScore, energyScore int, history []models.ChatMessage, userMessage string) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("[ai] failed to create gemini client: %v", err)
		return ChatFallbackMessage, nil
	}
	defer client.Close()

	model := client.GenerativeModel(geminiModel)
	model.SetTemperature(0.9)
	model.SystemInstruction = genai.NewUserContent(genai.Text(fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\nReply to your owner's message briefly (1-3 sentences), warmly, like a casual chat — not like a formal assistant.",
		miloPersona, miloSafetyGuidelines, buildMiloContext(moodState, healthScore, energyScore),
	)))

	cs := model.StartChat()
	cs.History = buildChatHistory(history)

	resp, err := cs.SendMessage(ctx, genai.Text(userMessage))
	if err != nil {
		log.Printf("[ai] gemini chat error: %v", err)
		return ChatFallbackMessage, nil
	}

	text, err := extractResponseText(resp)
	if err != nil {
		log.Printf("[ai] gemini chat response was empty: %v", err)
		return ChatFallbackMessage, nil
	}

	return text, nil
}

// buildChatHistory converts stored chat messages into Gemini's chat history
// format, mapping our "milo" role to Gemini's "model" role.
func buildChatHistory(history []models.ChatMessage) []*genai.Content {
	contents := make([]*genai.Content, 0, len(history))
	for _, m := range history {
		role := "user"
		if m.Role == models.ChatRoleMilo {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(m.Content)},
		})
	}
	return contents
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
