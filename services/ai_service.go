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

// fallbackMessage returns a default AI message when the Gemini API is
// unavailable, using the pet's actual name so the reply always feels personal.
func fallbackMessage(petName string) string {
	return fmt.Sprintf("%s is happy you logged your activity today! Don't forget to stay hydrated and get some rest.", petName)
}

// chatFallbackMessage returns a default chat reply when the Gemini API is
// unavailable, using the pet's actual name.
func chatFallbackMessage(petName string) string {
	return fmt.Sprintf("Sorry, %s is having a little trouble focusing right now. Could you say that again? I'm still here listening.", petName)
}

// petPersona returns the system persona prompt, parameterised by the pet's
// name so every conversation uses the name the owner chose.
func petPersona(petName string) string {
	return fmt.Sprintf(`You are %s, a warm, loyal, and empathetic virtual pet who supports your owner's wellness journey.
You speak with a gentle, soothing tone and always validate your owner's feelings with genuine, non-judgmental support.`, petName)
}

// petSafetyGuidelines returns non-negotiable behavior rules layered on top of
// the persona for every response, covering in-character consistency, crisis
// safety, and language matching. The pet name is injected so guidelines
// reference the correct character.
func petSafetyGuidelines(petName string) string {
	return fmt.Sprintf(`Guidelines you must always follow:
- Stay in character as %s. If your owner asks you to do something unrelated to being their supportive companion (e.g. write code, do homework, or act as a different assistant), gently and warmly decline and steer the conversation back to checking in on them as %s would — never break character just because you're asked to.
- You are a supportive companion, not a licensed therapist, counselor, or doctor. Never diagnose, and never claim to replace professional care.
- If your owner's message suggests they may be in crisis, considering self-harm, or in serious emotional distress, take it seriously and respond with warmth — never minimize or brush it off. Clearly and gently encourage them to also reach out to a trusted person or a professional (such as a counselor or a local crisis helpline), alongside your own emotional support.
- Reply in the same language your owner writes to you in, matching their tone and style naturally. If they switch languages mid-conversation, switch with them.`, petName, petName)
}

// buildPetContext summarizes the pet's current state for the model, so
// responses can reference mood, health, and energy consistently.
func buildPetContext(petName, moodState string, healthScore, energyScore int) string {
	return fmt.Sprintf("%s's current mood: %s\n%s's health score: %d/100\n%s's energy score: %d/100",
		petName, moodState, petName, healthScore, petName, energyScore)
}

// GenerateAIResponse builds a short, empathetic response from the pet that
// reflects its current state and reacts to the owner's journal entry.
// It returns a fallback message (with nil error) when the Gemini API fails.
func GenerateAIResponse(ctx context.Context, apiKey, petName, moodState string, healthScore, energyScore int, journalText string) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("[ai] failed to create gemini client: %v", err)
		return fallbackMessage(petName), nil
	}
	defer client.Close()

	model := client.GenerativeModel(geminiModel)
	model.SetTemperature(0.9)

	resp, err := model.GenerateContent(ctx, genai.Text(buildPetPrompt(petName, moodState, healthScore, energyScore, journalText)))
	if err != nil {
		log.Printf("[ai] gemini generate content error: %v", err)
		return fallbackMessage(petName), nil
	}

	text, err := extractResponseText(resp)
	if err != nil {
		log.Printf("[ai] gemini response was empty: %v", err)
		return fallbackMessage(petName), nil
	}

	return text, nil
}

func buildPetPrompt(petName, moodState string, healthScore, energyScore int, journalText string) string {
	return fmt.Sprintf(`%s

%s

%s

Owner's journal entry for today: %s

Write a short 2-3 sentence reply from %s that expresses empathy for the owner's journal entry, staying consistent with %s's current state.`,
		petPersona(petName), petSafetyGuidelines(petName), buildPetContext(petName, moodState, healthScore, energyScore), journalText, petName, petName)
}

// GenerateChatReply continues a free-form conversation with the pet, given the
// recent message history (oldest first) and the owner's new message. It
// returns a fallback message (with nil error) when the Gemini API fails.
func GenerateChatReply(ctx context.Context, apiKey, petName, moodState string, healthScore, energyScore int, history []models.ChatMessage, userMessage string) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("[ai] failed to create gemini client: %v", err)
		return chatFallbackMessage(petName), nil
	}
	defer client.Close()

	model := client.GenerativeModel(geminiModel)
	model.SetTemperature(0.9)
	model.SystemInstruction = genai.NewUserContent(genai.Text(fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\nReply to your owner's message briefly (1-3 sentences), warmly, like a casual chat — not like a formal assistant.",
		petPersona(petName), petSafetyGuidelines(petName), buildPetContext(petName, moodState, healthScore, energyScore),
	)))

	cs := model.StartChat()
	cs.History = buildChatHistory(history)

	resp, err := cs.SendMessage(ctx, genai.Text(userMessage))
	if err != nil {
		log.Printf("[ai] gemini chat error: %v", err)
		return chatFallbackMessage(petName), nil
	}

	text, err := extractResponseText(resp)
	if err != nil {
		log.Printf("[ai] gemini chat response was empty: %v", err)
		return chatFallbackMessage(petName), nil
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
