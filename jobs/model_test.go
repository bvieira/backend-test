package jobs

import "testing"

func TestJob_ID(t *testing.T) {
	tests := []struct {
		description string
		job         Job
		want        string
	}{
		{
			"simple case",
			Job{Title: "Tecnico de Informatica", Description: "some description", Salary: 3200, City: []string{"Joinville"}, CityFormatted: []string{"Joinville - SC (1)"}},
			"5d661133e37b6303720ecc9d3238e5a115b407fe",
		},
		{
			"with symbol",
			Job{Title: "Tecnico de Informatica!", Description: "some description", Salary: 3200, City: []string{"Joinville"}, CityFormatted: []string{"Joinville - SC (1)"}},
			"5d661133e37b6303720ecc9d3238e5a115b407fe",
		},
		{
			"with accent",
			Job{Title: "Técnico de Informática", Description: "some description", Salary: 3200, City: []string{"Joinville"}, CityFormatted: []string{"Joinville - SC (1)"}},
			"5d661133e37b6303720ecc9d3238e5a115b407fe",
		},
		{
			"with accent and symbol",
			Job{Title: "Técnico de Informática !", Description: "some description", Salary: 3200, City: []string{"Joinville"}, CityFormatted: []string{"Joinville - SC (1)"}},
			"5d661133e37b6303720ecc9d3238e5a115b407fe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			if got := tt.job.ID(); got != tt.want {
				t.Errorf("Job.ID() = %v, want %v", got, tt.want)
			}
		})
	}
}
