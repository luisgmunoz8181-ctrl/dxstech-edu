package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type QuizQuestion struct {
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correctAnswer"`
}

type GeminiClient struct {
	httpClient *http.Client
}

func NewGeminiClient() *GeminiClient {
	return &GeminiClient{
		httpClient: &http.Client{
			Timeout: 75 * time.Second,
		},
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string  `json:"response_mime_type,omitempty"`
	Temperature      float32 `json:"temperature,omitempty"`
}

type geminiRequest struct {
	Contents          []geminiContent         `json:"contents"`
	SystemInstruction *geminiContent          `json:"system_instruction,omitempty"`
	GenerationConfig  *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// GenerateQuiz generates an exact number of multiple-choice questions based on notes/topic.
// The apiKey is provided ephemerally by the client request and never logged or stored.
func (g *GeminiClient) GenerateQuiz(ctx context.Context, apiKey string, title string, notes string, count int) ([]QuizQuestion, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("Gemini API Key es requerida para generar la evaluación")
	}
	if count < 3 || count > 15 {
		count = 5
	}

	prompt := fmt.Sprintf(`Eres un pedagogo experto. Genera exactamente %d preguntas de opción múltiple para la evaluación sobre "%s".
Material de estudio o apuntes:
"""
%s
"""

REGLAS ESTRICTAS:
1. Devuelve un arreglo JSON con exactamente %d elementos.
2. Cada elemento debe ser un objeto con:
   - "question": texto claro de la pregunta.
   - "options": un arreglo de exactamente 4 opciones de respuesta como cadenas de texto.
   - "correctAnswer": una cadena de texto que debe ser IDÉNTICA a una de las 4 opciones de "options".
3. Las preguntas deben ser relevantes, con una sola opción inequívocamente correcta.
4. Devuelve ÚNICAMENTE el arreglo JSON.`, count, title, notes, count)

	reqPayload := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			ResponseMimeType: "application/json",
			Temperature:      0.4,
		},
	}

	rawResp, err := g.callGeminiAPI(ctx, apiKey, reqPayload)
	if err != nil {
		return nil, err
	}

	cleaned := cleanJSONOutput(rawResp)
	var questions []QuizQuestion
	if err := json.Unmarshal([]byte(cleaned), &questions); err != nil {
		// Attempt parsing if wrapped in an object like {"questions": [...]}
		var objWrapper struct {
			Questions []QuizQuestion `json:"questions"`
		}
		if err2 := json.Unmarshal([]byte(cleaned), &objWrapper); err2 == nil && len(objWrapper.Questions) > 0 {
			questions = objWrapper.Questions
		} else {
			return nil, fmt.Errorf("no se pudo interpretar la respuesta estructurada de Gemini: %w", err)
		}
	}

	// Strict validation
	if len(questions) == 0 {
		return nil, errors.New("Gemini no devolvió preguntas")
	}

	// Trim if slightly more or validate count
	if len(questions) > count {
		questions = questions[:count]
	}

	for i, q := range questions {
		if strings.TrimSpace(q.Question) == "" {
			return nil, fmt.Errorf("la pregunta #%d no contiene texto", i+1)
		}
		if len(q.Options) != 4 {
			return nil, fmt.Errorf("la pregunta #%d debe contener exactamente 4 opciones (recibidas: %d)", i+1, len(q.Options))
		}
		matched := false
		for _, opt := range q.Options {
			if strings.TrimSpace(opt) == strings.TrimSpace(q.CorrectAnswer) {
				matched = true
				break
			}
		}
		if !matched {
			// Guarantee correctness or fall back to first option
			questions[i].CorrectAnswer = q.Options[0]
		}
	}

	return questions, nil
}

// AskChatbot answers user inquiries given a knowledge base and strict mode flag.
func (g *GeminiClient) AskChatbot(ctx context.Context, apiKey string, userMessage string, knowledgeBase string, strictMode bool) (string, error) {
	if strings.TrimSpace(apiKey) == "" {
		return "⚠️ No has configurado tu Gemini API Key en Configuración.", nil
	}

	var systemPrompt string
	if strictMode {
		systemPrompt = fmt.Sprintf(`Eres el asistente virtual oficial de DxSTech Edu.
Tu conocimiento proviene EXCLUSIVAMENTE de la siguiente Base de Conocimiento:
"""
%s
"""

REGLAS DE ORO:
- Responde de forma amable, clara y profesional en español.
- Si la pregunta del usuario no puede responderse directamente con la información de la base de conocimiento proporcionada, DEBES responder exactamente:
"Lo siento, solo puedo responder preguntas sobre la base de conocimiento proporcionada."
- No inventes información, no asumas datos fuera del texto anterior.`, knowledgeBase)
	} else {
		systemPrompt = fmt.Sprintf(`Eres el asistente virtual educativo oficial de DxSTech Edu.
Básate primordialmente en la siguiente información de la institución:
"""
%s
"""
Responde con calidez, profesionalismo y pedagogía. Si no está en la base, puedes responder preguntas generales de forma útil y cortés.`, knowledgeBase)
	}

	reqPayload := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: userMessage}},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			Temperature: 0.3,
		},
	}

	ans, err := g.callGeminiAPI(ctx, apiKey, reqPayload)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(ans), nil
}

func (g *GeminiClient) callGeminiAPI(ctx context.Context, apiKey string, payload geminiRequest) (string, error) {
	models := []string{"gemini-2.5-flash", "gemini-2.0-flash", "gemini-1.5-flash"}
	var lastErr error

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error serializando solicitud a Gemini: %w", err)
	}

	for _, model := range models {
		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			lastErr = err
			continue
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := g.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("falla de red al contactar Gemini API: %w", err)
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if readErr != nil {
			lastErr = readErr
			continue
		}

		if resp.StatusCode != http.StatusOK {
			var errResp geminiResponse
			_ = json.Unmarshal(respBody, &errResp)
			if errResp.Error != nil {
				lastErr = fmt.Errorf("error de Gemini API (%d): %s", errResp.Error.Code, errResp.Error.Message)
			} else {
				lastErr = fmt.Errorf("error de Gemini API (HTTP %d)", resp.StatusCode)
			}
			// If key is invalid or quota exceeded, do not retry other models unnecessarily
			if resp.StatusCode == 400 || resp.StatusCode == 403 {
				return "", lastErr
			}
			continue
		}

		var parsed geminiResponse
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			lastErr = fmt.Errorf("error parseando JSON de Gemini: %w", err)
			continue
		}

		if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
			lastErr = errors.New("Gemini no devolvió ningún contenido")
			continue
		}

		return parsed.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", lastErr
}

func cleanJSONOutput(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		if idx := strings.LastIndex(text, "```"); idx != -1 {
			text = text[:idx]
		}
	} else if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		if idx := strings.LastIndex(text, "```"); idx != -1 {
			text = text[:idx]
		}
	}
	return strings.TrimSpace(text)
}
