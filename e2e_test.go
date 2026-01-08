package videocatalog

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/krelinga/video-catalog/vcrest"
	"github.com/oapi-codegen/nullable"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestVideoCatalogEnd2End(t *testing.T) {
	ctx := context.Background()

	serverURL := setup(t, ctx)
	client, err := vcrest.NewClientWithResponses(serverURL)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	t.Run("GET non-existing entities", func(t *testing.T) {
		testGetNonExisting(t, ctx, client)
	})

	t.Run("Work CRUD operations", func(t *testing.T) {
		testWorkCRUD(t, ctx, client)
	})

	t.Run("Source CRUD operations", func(t *testing.T) {
		testSourceCRUD(t, ctx, client)
	})

	t.Run("Plan CRUD operations", func(t *testing.T) {
		testPlanCRUD(t, ctx, client)
	})

	t.Run("ListPlans with multiple plans", func(t *testing.T) {
		testListPlans(t, ctx, client)
	})
}

func testGetNonExisting(t *testing.T, ctx context.Context, client *vcrest.ClientWithResponses) {
	nonExistingUUID := openapi_types.UUID(uuid.New())

	// Test GET on non-existing work
	workResp, err := client.GetWorkWithResponse(ctx, nonExistingUUID)
	if err != nil {
		t.Fatalf("GetWork failed: %v", err)
	}
	if workResp.StatusCode() != 404 {
		t.Errorf("Expected 404 for non-existing work, got %d", workResp.StatusCode())
	}

	// Test GET on non-existing source
	sourceResp, err := client.GetSourceWithResponse(ctx, nonExistingUUID)
	if err != nil {
		t.Fatalf("GetSource failed: %v", err)
	}
	if sourceResp.StatusCode() != 404 {
		t.Errorf("Expected 404 for non-existing source, got %d", sourceResp.StatusCode())
	}

	// Test GET on non-existing plan
	planResp, err := client.GetPlanWithResponse(ctx, nonExistingUUID)
	if err != nil {
		t.Fatalf("GetPlan failed: %v", err)
	}
	if planResp.StatusCode() != 404 {
		t.Errorf("Expected 404 for non-existing plan, got %d", planResp.StatusCode())
	}
}

func testWorkCRUD(t *testing.T, ctx context.Context, client *vcrest.ClientWithResponses) {
	// Test MovieWork
	t.Run("MovieWork", func(t *testing.T) {
		workUUID := openapi_types.UUID(uuid.New())

		// PUT MovieWork
		putResp, err := client.PutMovieWorkWithResponse(ctx, workUUID, vcrest.PutMovieWorkJSONRequestBody{
			Title:       nullable.NewNullableWithValue("The Matrix"),
			ReleaseYear: nullable.NewNullableWithValue(int32(1999)),
			TmdbId:      nullable.NewNullableWithValue(int32(603)),
		})
		if err != nil {
			t.Fatalf("PutMovieWork failed: %v", err)
		}
		if putResp.StatusCode() != 201 {
			t.Fatalf("Expected 201 for PUT, got %d: %s", putResp.StatusCode(), string(putResp.Body))
		}

		// GET MovieWork
		getResp, err := client.GetWorkWithResponse(ctx, workUUID)
		if err != nil {
			t.Fatalf("GetWork failed: %v", err)
		}
		if getResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for GET, got %d", getResp.StatusCode())
		}
		if getResp.JSON200 == nil || getResp.JSON200.Movie == nil {
			t.Fatal("Expected movie work in response")
		}
		if getResp.JSON200.Movie.Title.MustGet() != "The Matrix" {
			t.Errorf("Expected title 'The Matrix', got '%s'", getResp.JSON200.Movie.Title.MustGet())
		}
		if getResp.JSON200.Movie.ReleaseYear.MustGet() != 1999 {
			t.Errorf("Expected release year 1999, got %d", getResp.JSON200.Movie.ReleaseYear.MustGet())
		}

		// PATCH MovieWork
		patchResp, err := client.PatchMovieWorkWithResponse(ctx, workUUID, vcrest.PatchMovieWorkJSONRequestBody{
			ReleaseYear: nullable.NewNullableWithValue(int32(1998)),
			TmdbId:      nullable.NewNullableWithValue(int32(604)),
		})
		if err != nil {
			t.Fatalf("PatchMovieWork failed: %v", err)
		}
		if patchResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for PATCH, got %d: %s", patchResp.StatusCode(), string(patchResp.Body))
		}

		// GET again to verify PATCH
		getResp2, err := client.GetWorkWithResponse(ctx, workUUID)
		if err != nil {
			t.Fatalf("GetWork after PATCH failed: %v", err)
		}
		if getResp2.StatusCode() != 200 {
			t.Fatalf("Expected 200 for GET after PATCH, got %d", getResp2.StatusCode())
		}
		if getResp2.JSON200.Movie.Title.MustGet() != "The Matrix" {
			t.Errorf("Title should be unchanged, got '%s'", getResp2.JSON200.Movie.Title.MustGet())
		}
		if getResp2.JSON200.Movie.ReleaseYear.MustGet() != 1998 {
			t.Errorf("Expected updated release year 1998, got %d", getResp2.JSON200.Movie.ReleaseYear.MustGet())
		}
		if getResp2.JSON200.Movie.TmdbId.MustGet() != 604 {
			t.Errorf("Expected updated tmdbId 604, got %d", getResp2.JSON200.Movie.TmdbId.MustGet())
		}
	})

	// Test MovieEditionWork
	t.Run("MovieEditionWork", func(t *testing.T) {
		workUUID := openapi_types.UUID(uuid.New())

		// PUT MovieEdition
		putResp, err := client.PutMovieEditionWithResponse(ctx, workUUID, vcrest.PutMovieEditionJSONRequestBody{
			EditionType: nullable.NewNullableWithValue("Director's Cut"),
		})
		if err != nil {
			t.Fatalf("PutMovieEdition failed: %v", err)
		}
		if putResp.StatusCode() != 201 {
			t.Fatalf("Expected 201 for PUT, got %d", putResp.StatusCode())
		}

		// GET MovieEdition
		getResp, err := client.GetWorkWithResponse(ctx, workUUID)
		if err != nil {
			t.Fatalf("GetWork failed: %v", err)
		}
		if getResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for GET, got %d", getResp.StatusCode())
		}
		if getResp.JSON200 == nil || getResp.JSON200.MovieEdition == nil {
			t.Fatal("Expected movie edition work in response")
		}
		if getResp.JSON200.MovieEdition.EditionType.MustGet() != "Director's Cut" {
			t.Errorf("Expected edition type 'Director's Cut', got '%s'", getResp.JSON200.MovieEdition.EditionType.MustGet())
		}

		// PATCH MovieEdition
		patchResp, err := client.PatchMovieEditionWithResponse(ctx, workUUID, vcrest.PatchMovieEditionJSONRequestBody{
			EditionType: nullable.NewNullableWithValue("Extended Edition"),
		})
		if err != nil {
			t.Fatalf("PatchMovieEdition failed: %v", err)
		}
		if patchResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for PATCH, got %d", patchResp.StatusCode())
		}

		// GET again to verify PATCH
		getResp2, err := client.GetWorkWithResponse(ctx, workUUID)
		if err != nil {
			t.Fatalf("GetWork after PATCH failed: %v", err)
		}
		if getResp2.JSON200.MovieEdition.EditionType.MustGet() != "Extended Edition" {
			t.Errorf("Expected updated edition type 'Extended Edition', got '%s'", getResp2.JSON200.MovieEdition.EditionType.MustGet())
		}
	})
}

func testSourceCRUD(t *testing.T, ctx context.Context, client *vcrest.ClientWithResponses) {
	// Test FileSource
	t.Run("FileSource", func(t *testing.T) {
		sourceUUID := openapi_types.UUID(uuid.New())

		// PUT FileSource
		putResp, err := client.PutFileSourceWithResponse(ctx, sourceUUID, vcrest.PutFileSourceJSONRequestBody{
			Path: nullable.NewNullableWithValue("/media/movies/matrix.mkv"),
		})
		if err != nil {
			t.Fatalf("PutFileSource failed: %v", err)
		}
		if putResp.StatusCode() != 201 {
			t.Fatalf("Expected 201 for PUT, got %d", putResp.StatusCode())
		}

		// GET FileSource
		getResp, err := client.GetSourceWithResponse(ctx, sourceUUID)
		if err != nil {
			t.Fatalf("GetSource failed: %v", err)
		}
		if getResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for GET, got %d", getResp.StatusCode())
		}
		if getResp.JSON200 == nil || getResp.JSON200.File == nil {
			t.Fatal("Expected file source in response")
		}
		if getResp.JSON200.File.Path.MustGet() != "/media/movies/matrix.mkv" {
			t.Errorf("Expected path '/media/movies/matrix.mkv', got '%s'", getResp.JSON200.File.Path.MustGet())
		}

		// PATCH FileSource
		patchResp, err := client.PatchFileSourceWithResponse(ctx, sourceUUID, vcrest.PatchFileSourceJSONRequestBody{
			Path: nullable.NewNullableWithValue("/media/movies/matrix_remastered.mkv"),
		})
		if err != nil {
			t.Fatalf("PatchFileSource failed: %v", err)
		}
		if patchResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for PATCH, got %d", patchResp.StatusCode())
		}

		// GET again to verify PATCH
		getResp2, err := client.GetSourceWithResponse(ctx, sourceUUID)
		if err != nil {
			t.Fatalf("GetSource after PATCH failed: %v", err)
		}
		if getResp2.JSON200.File.Path.MustGet() != "/media/movies/matrix_remastered.mkv" {
			t.Errorf("Expected updated path, got '%s'", getResp2.JSON200.File.Path.MustGet())
		}
	})

	// Test DiscSource
	t.Run("DiscSource", func(t *testing.T) {
		sourceUUID := openapi_types.UUID(uuid.New())

		// PUT DiscSource
		putResp, err := client.PutDiscSourceWithResponse(ctx, sourceUUID, vcrest.PutDiscSourceJSONRequestBody{
			OrigDirName:   nullable.NewNullableWithValue("MATRIX_DISC"),
			Path:          nullable.NewNullableWithValue("/media/discs/matrix"),
			AllFilesAdded: nullable.NewNullableWithValue(false),
		})
		if err != nil {
			t.Fatalf("PutDiscSource failed: %v", err)
		}
		if putResp.StatusCode() != 201 {
			t.Fatalf("Expected 201 for PUT, got %d", putResp.StatusCode())
		}

		// GET DiscSource
		getResp, err := client.GetSourceWithResponse(ctx, sourceUUID)
		if err != nil {
			t.Fatalf("GetSource failed: %v", err)
		}
		if getResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for GET, got %d", getResp.StatusCode())
		}
		if getResp.JSON200 == nil || getResp.JSON200.Disc == nil {
			t.Fatal("Expected disc source in response")
		}
		if getResp.JSON200.Disc.OrigDirName.MustGet() != "MATRIX_DISC" {
			t.Errorf("Expected orig dir name 'MATRIX_DISC', got '%s'", getResp.JSON200.Disc.OrigDirName.MustGet())
		}

		// PATCH DiscSource
		patchResp, err := client.PatchDiscSourceWithResponse(ctx, sourceUUID, vcrest.PatchDiscSourceJSONRequestBody{
			AllFilesAdded: nullable.NewNullableWithValue(true),
		})
		if err != nil {
			t.Fatalf("PatchDiscSource failed: %v", err)
		}
		if patchResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for PATCH, got %d", patchResp.StatusCode())
		}

		// GET again to verify PATCH
		getResp2, err := client.GetSourceWithResponse(ctx, sourceUUID)
		if err != nil {
			t.Fatalf("GetSource after PATCH failed: %v", err)
		}
		if getResp2.JSON200.Disc.AllFilesAdded.MustGet() != true {
			t.Errorf("Expected AllFilesAdded to be true, got %v", getResp2.JSON200.Disc.AllFilesAdded.MustGet())
		}
		if getResp2.JSON200.Disc.OrigDirName.MustGet() != "MATRIX_DISC" {
			t.Errorf("OrigDirName should be unchanged, got '%s'", getResp2.JSON200.Disc.OrigDirName.MustGet())
		}
	})
}

func testPlanCRUD(t *testing.T, ctx context.Context, client *vcrest.ClientWithResponses) {
	// Create work and source for plan tests
	workUUID := openapi_types.UUID(uuid.New())
	sourceUUID := openapi_types.UUID(uuid.New())

	_, err := client.PutMovieWorkWithResponse(ctx, workUUID, vcrest.PutMovieWorkJSONRequestBody{
		Title: nullable.NewNullableWithValue("Test Movie"),
	})
	if err != nil {
		t.Fatalf("Failed to create work: %v", err)
	}

	_, err = client.PutFileSourceWithResponse(ctx, sourceUUID, vcrest.PutFileSourceJSONRequestBody{
		Path: nullable.NewNullableWithValue("/test/path.mkv"),
	})
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	// Test DirectPlan
	t.Run("DirectPlan", func(t *testing.T) {
		planUUID := openapi_types.UUID(uuid.New())

		// PUT DirectPlan
		putResp, err := client.PutDirectPlanWithResponse(ctx, planUUID, vcrest.PutDirectPlanJSONRequestBody{
			SourceUuid: nullable.NewNullableWithValue(sourceUUID),
			WorkUuid:   nullable.NewNullableWithValue(workUUID),
		})
		if err != nil {
			t.Fatalf("PutDirectPlan failed: %v", err)
		}
		if putResp.StatusCode() != 201 {
			t.Fatalf("Expected 201 for PUT, got %d: %s", putResp.StatusCode(), string(putResp.Body))
		}

		// GET DirectPlan
		getResp, err := client.GetPlanWithResponse(ctx, planUUID)
		if err != nil {
			t.Fatalf("GetPlan failed: %v", err)
		}
		if getResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for GET, got %d", getResp.StatusCode())
		}
		if getResp.JSON200 == nil || getResp.JSON200.Direct == nil {
			t.Fatal("Expected direct plan in response")
		}
		if openapi_types.UUID(getResp.JSON200.Direct.SourceUuid.MustGet()) != sourceUUID {
			t.Errorf("Expected source UUID %s, got %s", sourceUUID, getResp.JSON200.Direct.SourceUuid.MustGet())
		}

		// Create another source for PATCH test
		newSourceUUID := openapi_types.UUID(uuid.New())
		_, err = client.PutFileSourceWithResponse(ctx, newSourceUUID, vcrest.PutFileSourceJSONRequestBody{
			Path: nullable.NewNullableWithValue("/test/path2.mkv"),
		})
		if err != nil {
			t.Fatalf("Failed to create second source: %v", err)
		}

		// PATCH DirectPlan
		patchResp, err := client.PatchDirectPlanWithResponse(ctx, planUUID, vcrest.PatchDirectPlanJSONRequestBody{
			SourceUuid: nullable.NewNullableWithValue(newSourceUUID),
		})
		if err != nil {
			t.Fatalf("PatchDirectPlan failed: %v", err)
		}
		if patchResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for PATCH, got %d: %s", patchResp.StatusCode(), string(patchResp.Body))
		}

		// GET again to verify PATCH
		getResp2, err := client.GetPlanWithResponse(ctx, planUUID)
		if err != nil {
			t.Fatalf("GetPlan after PATCH failed: %v", err)
		}
		if openapi_types.UUID(getResp2.JSON200.Direct.SourceUuid.MustGet()) != newSourceUUID {
			t.Errorf("Expected updated source UUID %s, got %s", newSourceUUID, getResp2.JSON200.Direct.SourceUuid.MustGet())
		}
		if openapi_types.UUID(getResp2.JSON200.Direct.WorkUuid.MustGet()) != workUUID {
			t.Errorf("Work UUID should be unchanged, got %s", getResp2.JSON200.Direct.WorkUuid.MustGet())
		}
	})

	// Test ChapterRangePlan
	t.Run("ChapterRangePlan", func(t *testing.T) {
		planUUID := openapi_types.UUID(uuid.New())

		// PUT ChapterRangePlan
		putResp, err := client.PutChapterRangePlanWithResponse(ctx, planUUID, vcrest.PutChapterRangePlanJSONRequestBody{
			SourceUuid:   nullable.NewNullableWithValue(sourceUUID),
			WorkUuid:     nullable.NewNullableWithValue(workUUID),
			StartChapter: nullable.NewNullableWithValue(int32(1)),
			EndChapter:   nullable.NewNullableWithValue(int32(5)),
		})
		if err != nil {
			t.Fatalf("PutChapterRangePlan failed: %v", err)
		}
		if putResp.StatusCode() != 201 {
			t.Fatalf("Expected 201 for PUT, got %d", putResp.StatusCode())
		}

		// GET ChapterRangePlan
		getResp, err := client.GetPlanWithResponse(ctx, planUUID)
		if err != nil {
			t.Fatalf("GetPlan failed: %v", err)
		}
		if getResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for GET, got %d", getResp.StatusCode())
		}
		if getResp.JSON200 == nil || getResp.JSON200.ChapterRange == nil {
			t.Fatal("Expected chapter range plan in response")
		}
		if getResp.JSON200.ChapterRange.StartChapter.MustGet() != 1 {
			t.Errorf("Expected start chapter 1, got %d", getResp.JSON200.ChapterRange.StartChapter.MustGet())
		}
		if getResp.JSON200.ChapterRange.EndChapter.MustGet() != 5 {
			t.Errorf("Expected end chapter 5, got %d", getResp.JSON200.ChapterRange.EndChapter.MustGet())
		}

		// PATCH ChapterRangePlan
		patchResp, err := client.PatchChapterRangePlanWithResponse(ctx, planUUID, vcrest.PatchChapterRangePlanJSONRequestBody{
			StartChapter: nullable.NewNullableWithValue(int32(2)),
			EndChapter:   nullable.NewNullableWithValue(int32(10)),
		})
		if err != nil {
			t.Fatalf("PatchChapterRangePlan failed: %v", err)
		}
		if patchResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for PATCH, got %d", patchResp.StatusCode())
		}

		// GET again to verify PATCH
		getResp2, err := client.GetPlanWithResponse(ctx, planUUID)
		if err != nil {
			t.Fatalf("GetPlan after PATCH failed: %v", err)
		}
		if getResp2.JSON200.ChapterRange.StartChapter.MustGet() != 2 {
			t.Errorf("Expected updated start chapter 2, got %d", getResp2.JSON200.ChapterRange.StartChapter.MustGet())
		}
		if getResp2.JSON200.ChapterRange.EndChapter.MustGet() != 10 {
			t.Errorf("Expected updated end chapter 10, got %d", getResp2.JSON200.ChapterRange.EndChapter.MustGet())
		}
	})
}

func testListPlans(t *testing.T, ctx context.Context, client *vcrest.ClientWithResponses) {
	// Create works and sources
	work1UUID := openapi_types.UUID(uuid.New())
	work2UUID := openapi_types.UUID(uuid.New())
	source1UUID := openapi_types.UUID(uuid.New())
	source2UUID := openapi_types.UUID(uuid.New())

	_, err := client.PutMovieWorkWithResponse(ctx, work1UUID, vcrest.PutMovieWorkJSONRequestBody{
		Title: nullable.NewNullableWithValue("Movie 1"),
	})
	if err != nil {
		t.Fatalf("Failed to create work 1: %v", err)
	}

	_, err = client.PutMovieWorkWithResponse(ctx, work2UUID, vcrest.PutMovieWorkJSONRequestBody{
		Title: nullable.NewNullableWithValue("Movie 2"),
	})
	if err != nil {
		t.Fatalf("Failed to create work 2: %v", err)
	}

	_, err = client.PutFileSourceWithResponse(ctx, source1UUID, vcrest.PutFileSourceJSONRequestBody{
		Path: nullable.NewNullableWithValue("/media/file1.mkv"),
	})
	if err != nil {
		t.Fatalf("Failed to create source 1: %v", err)
	}

	_, err = client.PutFileSourceWithResponse(ctx, source2UUID, vcrest.PutFileSourceJSONRequestBody{
		Path: nullable.NewNullableWithValue("/media/file2.mkv"),
	})
	if err != nil {
		t.Fatalf("Failed to create source 2: %v", err)
	}

	// Create multiple plans
	plan1UUID := openapi_types.UUID(uuid.New())
	plan2UUID := openapi_types.UUID(uuid.New())
	plan3UUID := openapi_types.UUID(uuid.New())

	// Create direct plan
	_, err = client.PutDirectPlanWithResponse(ctx, plan1UUID, vcrest.PutDirectPlanJSONRequestBody{
		SourceUuid: nullable.NewNullableWithValue(source1UUID),
		WorkUuid:   nullable.NewNullableWithValue(work1UUID),
	})
	if err != nil {
		t.Fatalf("Failed to create plan 1: %v", err)
	}

	// Create chapter range plan
	_, err = client.PutChapterRangePlanWithResponse(ctx, plan2UUID, vcrest.PutChapterRangePlanJSONRequestBody{
		SourceUuid:   nullable.NewNullableWithValue(source2UUID),
		WorkUuid:     nullable.NewNullableWithValue(work2UUID),
		StartChapter: nullable.NewNullableWithValue(int32(1)),
		EndChapter:   nullable.NewNullableWithValue(int32(3)),
	})
	if err != nil {
		t.Fatalf("Failed to create plan 2: %v", err)
	}

	// Create another direct plan
	_, err = client.PutDirectPlanWithResponse(ctx, plan3UUID, vcrest.PutDirectPlanJSONRequestBody{
		SourceUuid: nullable.NewNullableWithValue(source1UUID),
		WorkUuid:   nullable.NewNullableWithValue(work2UUID),
	})
	if err != nil {
		t.Fatalf("Failed to create plan 3: %v", err)
	}

	// List all plans
	t.Run("ListAllPlans", func(t *testing.T) {
		listResp, err := client.ListPlansWithResponse(ctx, &vcrest.ListPlansParams{})
		if err != nil {
			t.Fatalf("ListPlans failed: %v", err)
		}
		if listResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for ListPlans, got %d", listResp.StatusCode())
		}
		if len(listResp.JSON200.Plans) < 3 {
			t.Errorf("Expected at least 3 plans, got %d", len(listResp.JSON200.Plans))
		}

		// Verify we have both plan types
		foundDirect := false
		foundChapterRange := false
		for _, plan := range listResp.JSON200.Plans {
			if plan.Direct != nil {
				foundDirect = true
			}
			if plan.ChapterRange != nil {
				foundChapterRange = true
			}
		}
		if !foundDirect {
			t.Error("Expected to find at least one direct plan")
		}
		if !foundChapterRange {
			t.Error("Expected to find at least one chapter range plan")
		}
	})

	// List plans filtered by source
	t.Run("ListPlansBySource", func(t *testing.T) {
		listResp, err := client.ListPlansWithResponse(ctx, &vcrest.ListPlansParams{
			SourceUuid: &source1UUID,
		})
		if err != nil {
			t.Fatalf("ListPlans with source filter failed: %v", err)
		}
		if listResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for filtered ListPlans, got %d", listResp.StatusCode())
		}
		if len(listResp.JSON200.Plans) != 2 {
			t.Errorf("Expected 2 plans for source1, got %d", len(listResp.JSON200.Plans))
		}

		// Verify all returned plans use the correct source
		for _, plan := range listResp.JSON200.Plans {
			var planSource openapi_types.UUID
			if plan.Direct != nil {
				planSource = plan.Direct.SourceUuid.MustGet()
			} else if plan.ChapterRange != nil {
				planSource = plan.ChapterRange.SourceUuid.MustGet()
			}
			if planSource != source1UUID {
				t.Errorf("Expected all plans to have source %s, got %s", source1UUID, planSource)
			}
		}
	})

	// List plans filtered by work
	t.Run("ListPlansByWork", func(t *testing.T) {
		listResp, err := client.ListPlansWithResponse(ctx, &vcrest.ListPlansParams{
			WorkUuid: &work2UUID,
		})
		if err != nil {
			t.Fatalf("ListPlans with work filter failed: %v", err)
		}
		if listResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for filtered ListPlans, got %d", listResp.StatusCode())
		}
		if len(listResp.JSON200.Plans) != 2 {
			t.Errorf("Expected 2 plans for work2, got %d", len(listResp.JSON200.Plans))
		}

		// Verify all returned plans use the correct work
		for _, plan := range listResp.JSON200.Plans {
			var planWork openapi_types.UUID
			if plan.Direct != nil {
				planWork = plan.Direct.WorkUuid.MustGet()
			} else if plan.ChapterRange != nil {
				planWork = plan.ChapterRange.WorkUuid.MustGet()
			}
			if planWork != work2UUID {
				t.Errorf("Expected all plans to have work %s, got %s", work2UUID, planWork)
			}
		}
	})

	// Test pagination
	t.Run("ListPlansWithPagination", func(t *testing.T) {
		pageSize := int32(2)
		listResp, err := client.ListPlansWithResponse(ctx, &vcrest.ListPlansParams{
			PageSize: &pageSize,
		})
		if err != nil {
			t.Fatalf("ListPlans with page size failed: %v", err)
		}
		if listResp.StatusCode() != 200 {
			t.Fatalf("Expected 200 for paginated ListPlans, got %d", listResp.StatusCode())
		}
		if len(listResp.JSON200.Plans) > 2 {
			t.Errorf("Expected at most 2 plans with page size 2, got %d", len(listResp.JSON200.Plans))
		}

		// If there's a next page token, fetch the next page
		if listResp.JSON200.NextPageToken != nil && *listResp.JSON200.NextPageToken != "" {
			listResp2, err := client.ListPlansWithResponse(ctx, &vcrest.ListPlansParams{
				PageSize:  &pageSize,
				PageToken: listResp.JSON200.NextPageToken,
			})
			if err != nil {
				t.Fatalf("ListPlans with page token failed: %v", err)
			}
			if listResp2.StatusCode() != 200 {
				t.Fatalf("Expected 200 for second page, got %d", listResp2.StatusCode())
			}
			// Verify we got different plans
			if len(listResp2.JSON200.Plans) > 0 && listResp.JSON200.Plans[0].Uuid == listResp2.JSON200.Plans[0].Uuid {
				t.Error("Expected different plans on second page")
			}
		}
	})
}

// Starts the server container and all dependencies, and returns a URL string that can be used in client connections.
func setup(t *testing.T, ctx context.Context) string {
	// Create docker network.
	net, err := network.New(ctx, network.WithCheckDuplicate())
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}
	networkName := net.Name

	// Database configuration
	const (
		dbHost = "postgres"
		dbPort = "5432"
		dbName = "videocatalog"
		dbUser = "videocataloguser"
		dbPass = "videocatalogpass"
	)

	// Start Postgres container
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPass,
		},
		Networks:       []string{networkName},
		NetworkAliases: map[string][]string{networkName: {dbHost}},
		WaitingFor:     wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		dumpContainerLogs(t, ctx, postgresContainer, dbHost)
	})

	// Build and start the server container
	serverReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    ".",
			Dockerfile: "Dockerfile",
			BuildArgs:  map[string]*string{},
		},
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"VC_SERVER_PORT": "8080",
			"VC_DB_HOST":     dbHost,
			"VC_DB_PORT":     dbPort,
			"VC_DB_NAME":     dbName,
			"VC_DB_USER":     dbUser,
			"VC_DB_PASSWORD": dbPass,
		},
		Networks:       []string{networkName},
		NetworkAliases: map[string][]string{networkName: {"server"}},
		WaitingFor:     wait.ForLog("Starting HTTP server on port 8080"),
	}
	serverContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: serverReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start server container: %v", err)
	}
	t.Cleanup(func() {
		dumpContainerLogs(t, ctx, serverContainer, "server")
	})

	// Get server mapped port
	mappedPort, err := serverContainer.MappedPort(ctx, "8080")
	if err != nil {
		t.Fatalf("failed to get server mapped port: %v", err)
	}
	serverHost, err := serverContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get server host: %v", err)
	}

	return fmt.Sprintf("http://%s:%s", serverHost, mappedPort.Port())
}

// dumpContainerLogs reads and logs all output from a container
func dumpContainerLogs(t *testing.T, ctx context.Context, container testcontainers.Container, name string) {
	logs, err := container.Logs(ctx)
	if err != nil {
		t.Logf("failed to get %s container logs: %v", name, err)
		return
	}
	defer logs.Close()

	logBytes, err := io.ReadAll(logs)
	if err != nil {
		t.Logf("failed to read %s container logs: %v", name, err)
		return
	}

	t.Logf("=== %s container logs ===\n%s", name, string(logBytes))
}
