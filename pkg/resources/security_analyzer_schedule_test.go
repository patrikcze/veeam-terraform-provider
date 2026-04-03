package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/patrikcze/terraform-provider-veeam/internal/client"
	"github.com/patrikcze/terraform-provider-veeam/internal/models"
)

func TestSecurityAnalyzerSchedule_Metadata(t *testing.T) {
	r := NewSecurityAnalyzerSchedule()
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "veeam"}, &resp)
	assert.Equal(t, "veeam_security_analyzer_schedule", resp.TypeName)
}

func TestSecurityAnalyzerSchedule_Schema(t *testing.T) {
	r := &SecurityAnalyzerSchedule{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	assert.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "run_automatically")
	assert.Contains(t, resp.Schema.Attributes, "daily_enabled")
	assert.Contains(t, resp.Schema.Attributes, "daily_local_time")
	assert.Contains(t, resp.Schema.Attributes, "monthly_enabled")
	assert.Contains(t, resp.Schema.Attributes, "monthly_day_of_month")
}

func TestSecurityAnalyzerSchedule_Configure_Nil(t *testing.T) {
	r := &SecurityAnalyzerSchedule{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestSecurityAnalyzerSchedule_Configure_InvalidType(t *testing.T) {
	r := &SecurityAnalyzerSchedule{}
	req := resource.ConfigureRequest{ProviderData: "bad"}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	assert.True(t, resp.Diagnostics.HasError())
}

func TestSecurityAnalyzerSchedule_SyncFromAPI_FullResponse(t *testing.T) {
	data := &SecurityAnalyzerScheduleModel{}
	api := &models.SecurityAnalyzerScheduleModel{
		RunAutomatically: true,
		Daily: &models.SecurityAnalyzerDailySchedule{
			IsEnabled: true,
			LocalTime: "02:00",
		},
		Monthly: &models.SecurityAnalyzerMonthlySchedule{
			IsEnabled:  false,
			DayOfMonth: 15,
		},
	}

	syncSecurityAnalyzerScheduleFromAPI(data, api)

	assert.True(t, data.RunAutomatically.ValueBool())
	assert.True(t, data.DailyEnabled.ValueBool())
	assert.Equal(t, "02:00", data.DailyLocalTime.ValueString())
	assert.False(t, data.MonthlyEnabled.ValueBool())
	assert.Equal(t, int64(15), data.MonthlyDayOfMonth.ValueInt64())
}

func TestSecurityAnalyzerSchedule_SyncFromAPI_NilSubobjects(t *testing.T) {
	data := &SecurityAnalyzerScheduleModel{}
	api := &models.SecurityAnalyzerScheduleModel{
		RunAutomatically: false,
		Daily:            nil,
		Monthly:          nil,
	}

	syncSecurityAnalyzerScheduleFromAPI(data, api)

	assert.False(t, data.RunAutomatically.ValueBool())
	assert.False(t, data.DailyEnabled.ValueBool())
	assert.False(t, data.MonthlyEnabled.ValueBool())
}

func TestSecurityAnalyzerSchedule_Create_Success(t *testing.T) {
	mockClient := new(MockVeeamClient)

	// GET current state, then PUT updated state
	mockClient.On("GetJSON", mock.Anything, client.PathSecurityAnalyzerSchedule, mock.Anything).
		Run(func(args mock.Arguments) {
			result := args.Get(2).(*models.SecurityAnalyzerScheduleModel)
			result.RunAutomatically = false
		}).Return(nil)

	mockClient.On("PutJSON", mock.Anything, client.PathSecurityAnalyzerSchedule, mock.Anything, nil).
		Return(nil)

	r := &SecurityAnalyzerSchedule{client: mockClient}
	data := &SecurityAnalyzerScheduleModel{
		RunAutomatically:  types.BoolValue(true),
		DailyEnabled:      types.BoolNull(),
		DailyLocalTime:    types.StringNull(),
		MonthlyEnabled:    types.BoolNull(),
		MonthlyDayOfMonth: types.Int64Null(),
	}

	err := r.applySchedule(context.Background(), data)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSecurityAnalyzerSchedule_Create_GetFails(t *testing.T) {
	mockClient := new(MockVeeamClient)

	mockClient.On("GetJSON", mock.Anything, client.PathSecurityAnalyzerSchedule, mock.Anything).
		Return(fmt.Errorf("HTTP 500"))

	r := &SecurityAnalyzerSchedule{client: mockClient}
	data := &SecurityAnalyzerScheduleModel{
		RunAutomatically: types.BoolValue(true),
	}

	err := r.applySchedule(context.Background(), data)
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestSecurityAnalyzerSchedule_Delete_NoOp(t *testing.T) {
	r := &SecurityAnalyzerSchedule{}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), resource.DeleteRequest{}, resp)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestSecurityAnalyzerSchedule_ImportState(t *testing.T) {
	var _ resource.ResourceWithImportState = &SecurityAnalyzerSchedule{}
}
