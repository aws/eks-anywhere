package docker_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/docker"
)

func TestConcurrentImageProcessorProcessSuccess(t *testing.T) {
	tests := []struct {
		name        string
		images      []string
		maxRoutines int
	}{
		{
			name:        "more jobs than routines",
			images:      []string{"image1:1", "image2:2", "images3:3"},
			maxRoutines: 2,
		},
		{
			name:        "same jobs than routines",
			images:      []string{"image1:1", "image2:2", "images3:3"},
			maxRoutines: 3,
		},
		{
			name:        "less jobs than routines",
			images:      []string{"image1:1", "image2:2", "images3:3"},
			maxRoutines: 4,
		},
		{
			name:        "zero routines",
			images:      []string{"image1:1", "image2:2", "images3:3"},
			maxRoutines: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()

			processor := docker.NewConcurrentImageProcessor(tt.maxRoutines)

			process := func(_ context.Context, _ string) error {
				return nil
			}

			g.Expect(processor.Process(ctx, tt.images, process)).To(Succeed())
		})
	}
}

func TestConcurrentImageProcessorProcessError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	images := []string{"image1:1", "image2:2", "images3:3"}

	processor := docker.NewConcurrentImageProcessor(2)

	process := func(_ context.Context, i string) error {
		if i == "image2:2" {
			return errors.New("processing error")
		}
		return nil
	}

	g.Expect(processor.Process(ctx, images, process)).To(
		MatchError(ContainSubstring("image processor worker failed, rest of jobs were aborted: processing error")),
	)
}

func TestConcurrentImageProcessorProcessErrorWithJobsBeingCancelled(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	images := []string{"image1:1", "image2:2", "images3:3"}

	processor := docker.NewConcurrentImageProcessor(2)

	process := func(ctx context.Context, i string) error {
		if i == "image2:2" {
			return errors.New("processing error")
		}
		// Block until context gets cancelled to trigger the flow
		// where jobs get cancelled after first error
		<-ctx.Done()
		return nil
	}

	g.Expect(processor.Process(ctx, images, process)).To(
		MatchError(ContainSubstring("image processor worker failed, rest of jobs were aborted: processing error")),
	)
}

func TestConcurrentImageProcessorProcessCancelParentContext(t *testing.T) {
	g := NewWithT(t)
	ctx, cancel := context.WithCancel(context.Background())
	images := []string{"image1:1", "image2:2", "images3:3"}

	processor := docker.NewConcurrentImageProcessor(2)

	process := func(ctx context.Context, i string) error {
		<-ctx.Done()
		return nil
	}

	cancel()

	g.Expect(processor.Process(ctx, images, process)).To(Succeed())
}
