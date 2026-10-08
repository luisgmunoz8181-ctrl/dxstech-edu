package enrollments

import (
	"time"
)

type Enrollment struct {
	ID                  string     `json:"id"`
	UserID              string     `json:"userId"`
	CourseID            string     `json:"courseId"`
	Status              string     `json:"status"` // 'active', 'completed', 'dropped'
	ProgressPercent     float64    `json:"progressPercent"`
	EnrolledAt          time.Time  `json:"enrolledAt"`
	CompletedAt         *time.Time `json:"completedAt,omitempty"`
	LastAccessedAt      time.Time  `json:"lastAccessedAt"`
	LastLessonID        string     `json:"lastLessonId,omitempty"`
	CourseTitle         string     `json:"courseTitle,omitempty"`
	CourseCode          string     `json:"courseCode,omitempty"`
	CourseThumbnail     string     `json:"courseThumbnail,omitempty"`
	CourseCategory      string     `json:"courseCategory,omitempty"`
	CourseLevel         string     `json:"courseLevel,omitempty"`
	CourseInstructor    string     `json:"courseInstructor,omitempty"`
	CourseDurationHours float64    `json:"courseDurationHours,omitempty"`
	CompletedLessons    int        `json:"completedLessons"`
	TotalLessons        int        `json:"totalLessons"`
	StudentName         string     `json:"studentName,omitempty"`
	StudentEmail        string     `json:"studentEmail,omitempty"`
}

type LessonProgress struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	CourseID    string    `json:"courseId"`
	LessonID    string    `json:"lessonId"`
	Completed   bool      `json:"completed"`
	CompletedAt time.Time `json:"completedAt"`
}

type StudentStats struct {
	TotalEnrolled   int     `json:"totalEnrolled"`
	InProgress      int     `json:"inProgress"`
	Completed       int     `json:"completed"`
	TotalHours      float64 `json:"totalHours"`
	AverageProgress float64 `json:"averageProgress"`
}

type ToggleProgressRequest struct {
	CourseID string `json:"courseId" binding:"required"`
	LessonID string `json:"lessonId" binding:"required"`
	Completed bool   `json:"completed"`
}

type AdminEnrollRequest struct {
	UserID   string `json:"userId" binding:"required"`
	CourseID string `json:"courseId" binding:"required"`
}
