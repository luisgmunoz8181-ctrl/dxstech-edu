package analytics

import "time"

// OverviewMetrics resumen general de KPIs de la plataforma LMS
type OverviewMetrics struct {
	TotalUsers           int     `json:"totalUsers"`
	ActiveUsers          int     `json:"activeUsers"`
	InactiveUsers        int     `json:"inactiveUsers"`
	StudentsCount        int     `json:"studentsCount"`
	InstructorsCount     int     `json:"instructorsCount"`
	TotalCourses         int     `json:"totalCourses"`
	PublishedCourses     int     `json:"publishedCourses"`
	DraftCourses         int     `json:"draftCourses"`
	ArchivedCourses      int     `json:"archivedCourses"`
	TotalEnrollments     int     `json:"totalEnrollments"`
	ActiveEnrollments    int     `json:"activeEnrollments"`
	CompletedEnrollments int     `json:"completedEnrollments"`
	CompletionRate       float64 `json:"completionRate"`
	AverageProgress      float64 `json:"averageProgress"`
	TotalCertificates    int     `json:"totalCertificates"`
	TotalHoursDelivered  float64 `json:"totalHoursDelivered"`
}

// CourseMetric analítica granular por cada curso
type CourseMetric struct {
	CourseID          string  `json:"courseId"`
	CourseCode        string  `json:"courseCode"`
	CourseTitle       string  `json:"courseTitle"`
	Category          string  `json:"category"`
	Status            string  `json:"status"`
	DurationHours     float64 `json:"durationHours"`
	TotalStudents     int     `json:"totalStudents"`
	ActiveStudents    int     `json:"activeStudents"`
	CompletedStudents int     `json:"completedStudents"`
	AverageProgress   float64 `json:"averageProgress"`
	CompletionRate    float64 `json:"completionRate"`
}

// TimelinePoint registro mensual para gráficos de tendencias
type TimelinePoint struct {
	Period string `json:"period"` // e.g. "2026-10"
	Count  int    `json:"count"`
}

// StudentMonitoringRow detalle de estudiante matriculado para la tabla administrativa
type StudentMonitoringRow struct {
	EnrollmentID    string     `json:"enrollmentId"`
	UserID          string     `json:"userId"`
	StudentName     string     `json:"studentName"`
	StudentEmail    string     `json:"studentEmail"`
	CourseID        string     `json:"courseId"`
	CourseCode      string     `json:"courseCode"`
	CourseTitle     string     `json:"courseTitle"`
	Category        string     `json:"category"`
	ProgressPercent float64    `json:"progressPercent"`
	Status          string     `json:"status"`
	CertificateID   string     `json:"certificateId,omitempty"`
	EnrolledAt      time.Time  `json:"enrolledAt"`
	CompletedAt     *time.Time `json:"completedAt,omitempty"`
	LastAccessedAt  time.Time  `json:"lastAccessedAt"`
}

// FilterOptions parámetros de filtrado para métricas y monitoreo
type FilterOptions struct {
	CourseID  string
	Category  string
	Status    string
	Search    string
	StartDate string
	EndDate   string
	Limit     int
	Offset    int
}
