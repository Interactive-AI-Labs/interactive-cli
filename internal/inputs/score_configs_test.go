package inputs

import "testing"

func TestBuildScoreConfigCreateBodyCategories(t *testing.T) {
	tests := []struct {
		name       string
		input      ScoreConfigCreateInput
		wantErr    bool
		errMessage string
	}{
		{
			name: "accepts array categories",
			input: ScoreConfigCreateInput{
				Name:           "relevance",
				DataType:       "categorical",
				CategoriesJSON: `["good","bad"]`,
			},
		},
		{
			name: "rejects object categories",
			input: ScoreConfigCreateInput{
				Name:           "relevance",
				DataType:       "categorical",
				CategoriesJSON: `{"bad":true}`,
			},
			wantErr:    true,
			errMessage: "invalid --categories: must be a valid JSON array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildScoreConfigCreateBody(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildScoreConfigCreateBody() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err.Error() != tt.errMessage {
				t.Fatalf("BuildScoreConfigCreateBody() error = %q, want %q", err.Error(), tt.errMessage)
			}
		})
	}
}

func TestBuildScoreConfigUpdateBodyCategories(t *testing.T) {
	tests := []struct {
		name       string
		input      ScoreConfigUpdateInput
		wantErr    bool
		errMessage string
	}{
		{
			name: "accepts array categories",
			input: ScoreConfigUpdateInput{
				CategoriesJSON: `["good","bad"]`,
			},
		},
		{
			name: "rejects scalar categories",
			input: ScoreConfigUpdateInput{
				CategoriesJSON: `"bad"`,
			},
			wantErr:    true,
			errMessage: "invalid --categories: must be a valid JSON array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildScoreConfigUpdateBody(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("BuildScoreConfigUpdateBody() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err.Error() != tt.errMessage {
				t.Fatalf("BuildScoreConfigUpdateBody() error = %q, want %q", err.Error(), tt.errMessage)
			}
		})
	}
}
