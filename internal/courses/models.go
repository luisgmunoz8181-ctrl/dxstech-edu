package courses

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type Course struct {
	ID                 string         `json:"id"`
	Title              string         `json:"title"`
	Code               string         `json:"code"`
	Slug               string         `json:"slug"`
	ShortDescription   string         `json:"shortDescription"`
	Description        string         `json:"description"`
	ThumbnailURL       string         `json:"thumbnailUrl"`
	Category           string         `json:"category"`
	InstructorName     string         `json:"instructorName"`
	DurationHours      float64        `json:"durationHours"`
	Level              string         `json:"level"`
	Status             string         `json:"status"` // 'draft', 'published', 'archived'
	PublishedAt        *time.Time     `json:"publishedAt,omitempty"`
	Requirements       string         `json:"requirements"`
	LearningObjectives string         `json:"learningObjectives"`
	CreatedBy          string         `json:"createdBy,omitempty"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
	ModulesCount       int            `json:"modulesCount"`
	LessonsCount       int            `json:"lessonsCount"`
	Modules            []CourseModule `json:"modules,omitempty"`
}

type CourseModule struct {
	ID          string    `json:"id"`
	CourseID    string    `json:"courseId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	OrderIndex  int       `json:"orderIndex"`
	Lessons     []Lesson  `json:"lessons,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Lesson struct {
	ID              string    `json:"id"`
	ModuleID        string    `json:"moduleId"`
	CourseID        string    `json:"courseId"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	ContentType     string    `json:"contentType"` // 'pdf', 'pptx', 'mp4', 'youtube', 'text', 'quiz'
	ContentURL      string    `json:"contentUrl"`
	ContentBody     string    `json:"contentBody"`
	QuizID          string    `json:"quizId,omitempty"`
	DurationMinutes int       `json:"durationMinutes"`
	OrderIndex      int       `json:"orderIndex"`
	IsFreePreview   bool      `json:"isFreePreview"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CreateCourseRequest struct {
	Title              string  `json:"title" binding:"required"`
	Code               string  `json:"code" binding:"required"`
	ShortDescription   string  `json:"shortDescription"`
	Description        string  `json:"description"`
	ThumbnailURL       string  `json:"thumbnailUrl"`
	Category           string  `json:"category"`
	InstructorName     string  `json:"instructorName"`
	DurationHours      float64 `json:"durationHours"`
	Level              string  `json:"level"`
	Requirements       string  `json:"requirements"`
	LearningObjectives string  `json:"learningObjectives"`
}

type UpdateCourseRequest struct {
	Title              string  `json:"title" binding:"required"`
	Code               string  `json:"code" binding:"required"`
	ShortDescription   string  `json:"shortDescription"`
	Description        string  `json:"description"`
	ThumbnailURL       string  `json:"thumbnailUrl"`
	Category           string  `json:"category"`
	InstructorName     string  `json:"instructorName"`
	DurationHours      float64 `json:"durationHours"`
	Level              string  `json:"level"`
	Requirements       string  `json:"requirements"`
	LearningObjectives string  `json:"learningObjectives"`
	Status             string  `json:"status"`
}

type ChangeStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type CreateModuleRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	OrderIndex  int    `json:"orderIndex"`
}

type UpdateModuleRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	OrderIndex  int    `json:"orderIndex"`
}

type CreateLessonRequest struct {
	Title           string `json:"title" binding:"required"`
	Description     string `json:"description"`
	ContentType     string `json:"contentType" binding:"required"`
	ContentURL      string `json:"contentUrl"`
	ContentBody     string `json:"contentBody"`
	QuizID          string `json:"quizId"`
	DurationMinutes int    `json:"durationMinutes"`
	OrderIndex      int    `json:"orderIndex"`
	IsFreePreview   bool   `json:"isFreePreview"`
}

type UpdateLessonRequest struct {
	Title           string `json:"title" binding:"required"`
	Description     string `json:"description"`
	ContentType     string `json:"contentType" binding:"required"`
	ContentURL      string `json:"contentUrl"`
	ContentBody     string `json:"contentBody"`
	QuizID          string `json:"quizId"`
	DurationMinutes int    `json:"durationMinutes"`
	OrderIndex      int    `json:"orderIndex"`
	IsFreePreview   bool   `json:"isFreePreview"`
}

type CourseDiscussion struct {
	ID        string             `json:"id"`
	CourseID  string             `json:"courseId"`
	LessonID  string             `json:"lessonId,omitempty"`
	UserID    string             `json:"userId"`
	UserName  string             `json:"userName"`
	UserRole  string             `json:"userRole"`
	Title     string             `json:"title"`
	Message   string             `json:"message"`
	ParentID  string             `json:"parentId,omitempty"`
	Replies   []CourseDiscussion `json:"replies"`
	CreatedAt time.Time          `json:"createdAt"`
}

type CreateDiscussionRequest struct {
	LessonID string `json:"lessonId"`
	Title    string `json:"title"`
	ParentID string `json:"parentId"`
	Message  string `json:"message" binding:"required"`
}

type CourseReview struct {
	ID        string    `json:"id"`
	CourseID  string    `json:"courseId"`
	UserID    string    `json:"userId"`
	UserName  string    `json:"userName"`
	Rating    int       `json:"rating" binding:"required"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required"`
	Comment string `json:"comment"`
}

type CourseReviewsSummary struct {
	AverageRating float64        `json:"averageRating"`
	TotalReviews  int            `json:"totalReviews"`
	Reviews       []CourseReview `json:"reviews"`
}

var nonSlugRegex = regexp.MustCompile(`[^a-z0-9]+`)

func GenerateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, "á", "a")
	slug = strings.ReplaceAll(slug, "é", "e")
	slug = strings.ReplaceAll(slug, "í", "i")
	slug = strings.ReplaceAll(slug, "ó", "o")
	slug = strings.ReplaceAll(slug, "ú", "u")
	slug = strings.ReplaceAll(slug, "ñ", "n")
	slug = nonSlugRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "curso"
	}
	return slug
}

func ValidateContentType(ct string) error {
	switch ct {
	case "pdf", "pptx", "mp4", "youtube", "text", "quiz":
		return nil
	default:
		return errors.New("tipo de contenido inválido: debe ser 'pdf', 'pptx', 'mp4', 'youtube', 'text' o 'quiz'")
	}
}
