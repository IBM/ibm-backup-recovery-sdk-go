/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package tasks

import (
	"testing"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
	"github.com/stretchr/testify/assert"
)

func TestBuildRetryOptions(t *testing.T) {
	t.Run("uses defaults when retry count is zero", func(t *testing.T) {
		result := buildRetryOptions(types.RetryOption{})

		assert.NotNil(t, result)
		assert.NotNil(t, result.Retries)
		assert.NotNil(t, result.RetryIntervalMins)
		assert.Equal(t, types.DefaultPolicyRetries, *result.Retries)
		assert.Equal(t, types.DefaultPolicyInterval, *result.RetryIntervalMins)
	})

	t.Run("uses provided retry settings", func(t *testing.T) {
		result := buildRetryOptions(types.RetryOption{
			NumberOfRetry:       5,
			WaitInCaseOfFailure: 15,
		})

		assert.NotNil(t, result)
		assert.Equal(t, int64(5), *result.Retries)
		assert.Equal(t, int64(15), *result.RetryIntervalMins)
	})
}

func TestBuildIncrementalSchedule(t *testing.T) {
	t.Run("returns error when unit is missing", func(t *testing.T) {
		result, err := buildIncrementalSchedule(&types.IncrementalBackup{})

		assert.Nil(t, result)
		assert.EqualError(t, err, "`Unit` is a required field for incrementalBackup")
	})

	t.Run("builds day schedule", func(t *testing.T) {
		unit := types.IncrementalBackup_Unit_Days
		result, err := buildIncrementalSchedule(&types.IncrementalBackup{
			Unit:        &unit,
			DaySchedule: &types.UnitDaySchedule{Every: 2},
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, string(unit), *result.Schedule.Unit)
		assert.NotNil(t, result.Schedule.DaySchedule)
		assert.Equal(t, int64(2), *result.Schedule.DaySchedule.Frequency)
	})

	t.Run("returns error when schedule for selected unit is missing", func(t *testing.T) {
		unit := types.IncrementalBackup_Unit_Hours
		result, err := buildIncrementalSchedule(&types.IncrementalBackup{
			Unit: &unit,
		})

		assert.Nil(t, result)
		assert.EqualError(t, err, "Hour schedule needs to be set for Unit set to Hours")
	})
}

func TestGenerateIncrementalSchedule(t *testing.T) {
	t.Run("builds schedules for all supported units", func(t *testing.T) {
		dayPolicy, err := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Days, &types.IncrementalBackup{
			DaySchedule: &types.UnitDaySchedule{Every: 1},
		})
		assert.NoError(t, err)
		assert.Equal(t, "Days", *dayPolicy.Schedule.Unit)
		assert.Equal(t, int64(1), *dayPolicy.Schedule.DaySchedule.Frequency)

		hourPolicy, err := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Hours, &types.IncrementalBackup{
			HourSchedule: &types.UnitDaySchedule{Every: 6},
		})
		assert.NoError(t, err)
		assert.Equal(t, "Hours", *hourPolicy.Schedule.Unit)
		assert.Equal(t, int64(6), *hourPolicy.Schedule.HourSchedule.Frequency)

		minutePolicy, err := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Minutes, &types.IncrementalBackup{
			MinuteSchedule: &types.UnitDaySchedule{Every: 30},
		})
		assert.NoError(t, err)
		assert.Equal(t, "Minutes", *minutePolicy.Schedule.Unit)
		assert.Equal(t, int64(30), *minutePolicy.Schedule.MinuteSchedule.Frequency)

		monthPolicy, err := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Months, &types.IncrementalBackup{
			MonthSchedule: &types.MonthSchedule{OnDate: 10},
		})
		assert.NoError(t, err)
		assert.Equal(t, "Months", *monthPolicy.Schedule.Unit)
		assert.Equal(t, int64(10), *monthPolicy.Schedule.MonthSchedule.DayOfMonth)

		weekPolicy, err := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Weeks, &types.IncrementalBackup{
			WeekSchedule: &types.WeekSchedule{OnDayOfWeek: []types.DayOfWeek{types.WeekDay_Monday, types.WeekDay_Friday}},
		})
		assert.NoError(t, err)
		assert.Equal(t, "Weeks", *weekPolicy.Schedule.Unit)
		assert.Equal(t, []string{"Monday", "Friday"}, weekPolicy.Schedule.WeekSchedule.DayOfWeek)

		yearDay := types.YearDay_First
		yearPolicy, err := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Years, &types.IncrementalBackup{
			YearSchedule: &types.YearSchedule{OnDayOfYear: &yearDay},
		})
		assert.NoError(t, err)
		assert.Equal(t, "Years", *yearPolicy.Schedule.Unit)
		assert.Equal(t, "First", *yearPolicy.Schedule.YearSchedule.DayOfYear)
	})

	t.Run("returns error when day schedule payload is missing", func(t *testing.T) {
		result, err := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Days, &types.IncrementalBackup{})

		assert.Nil(t, result)
		assert.EqualError(t, err, "Day schedule needs to be set for Unit set to Days")
	})

	t.Run("returns error for unsupported unit", func(t *testing.T) {
		result, err := GenerateIncrementalSchedule(types.IncrementalBackup_Unit("Invalid"), &types.IncrementalBackup{})

		assert.Nil(t, result)
		assert.EqualError(t, err, "Invalid incremental schedule unit provided")
	})

	t.Run("returns unit-specific errors for missing schedule payloads", func(t *testing.T) {
		hourResult, hourErr := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Hours, &types.IncrementalBackup{})
		assert.Nil(t, hourResult)
		assert.EqualError(t, hourErr, "Hour schedule needs to be set for Unit set to Hours")

		minuteResult, minuteErr := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Minutes, &types.IncrementalBackup{})
		assert.Nil(t, minuteResult)
		assert.EqualError(t, minuteErr, "Minute schedule needs to be set for Unit set to Minute")

		monthResult, monthErr := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Months, &types.IncrementalBackup{})
		assert.Nil(t, monthResult)
		assert.EqualError(t, monthErr, "Month schedule needs to be set for Unit set to Month")

		weekResult, weekErr := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Weeks, &types.IncrementalBackup{})
		assert.Nil(t, weekResult)
		assert.EqualError(t, weekErr, "Week schedule needs to be set for Unit set to Week")

		yearResult, yearErr := GenerateIncrementalSchedule(types.IncrementalBackup_Unit_Years, &types.IncrementalBackup{})
		assert.Nil(t, yearResult)
		assert.EqualError(t, yearErr, "Year schedule needs to be set for Unit set to Year")
	})
}

func TestBuildFullSchedule(t *testing.T) {
	t.Run("returns nil for empty input", func(t *testing.T) {
		result, err := buildFullSchedule(nil)

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when required fields are missing", func(t *testing.T) {
		result, err := buildFullSchedule([]types.FullBackup{{}})

		assert.Nil(t, result)
		assert.EqualError(t, err, "`Unit` and `Retention` are required fields for a FullBackup")
	})

	t.Run("builds schedules with retention", func(t *testing.T) {
		unit := types.FullSchedule_Unit_Weeks
		retentionUnit := types.Retention_Unit_Months
		result, err := buildFullSchedule([]types.FullBackup{
			{
				Unit:         &unit,
				WeekSchedule: &types.WeekSchedule{OnDayOfWeek: []types.DayOfWeek{types.WeekDay_Tuesday, types.WeekDay_Saturday}},
				Retention: &types.DataRetention{
					Unit:      &retentionUnit,
					RetainFor: 3,
				},
			},
		})

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "Weeks", *result[0].Schedule.Unit)
		assert.Equal(t, []string{"Tuesday", "Saturday"}, result[0].Schedule.WeekSchedule.DayOfWeek)
		assert.NotNil(t, result[0].Retention)
		assert.Equal(t, "Months", *result[0].Retention.Unit)
		assert.Equal(t, int64(3), *result[0].Retention.Duration)
	})

	t.Run("returns error from schedule or retention builders", func(t *testing.T) {
		dayUnit := types.FullSchedule_Unit_Days
		yearUnit := types.FullSchedule_Unit_Years
		badRetentionUnit := types.Retention_Unit_Weeks

		scheduleResult, scheduleErr := buildFullSchedule([]types.FullBackup{
			{
				Unit: &dayUnit,
				Retention: &types.DataRetention{
					Unit:      &badRetentionUnit,
					RetainFor: 1,
				},
			},
		})
		assert.Nil(t, scheduleResult)
		assert.EqualError(t, scheduleErr, "Day schedule needs to be set for Unit set to Days")

		retentionResult, retentionErr := buildFullSchedule([]types.FullBackup{
			{
				Unit:         &yearUnit,
				YearSchedule: &types.YearSchedule{OnDayOfYear: nil},
				Retention:    &types.DataRetention{},
			},
		})
		assert.Nil(t, retentionResult)
		assert.EqualError(t, retentionErr, " `Unit` and `RetainFor` are required parameters")
	})
}

func TestGenerateFullSchedule(t *testing.T) {
	t.Run("builds schedules for all supported units", func(t *testing.T) {
		dayPolicy, err := GenerateFullSchedule(types.FullSchedule_Unit_Days, &types.FullBackup{
			DaySchedule: &types.UnitDaySchedule{Every: 7},
		})
		assert.NoError(t, err)
		assert.Equal(t, "Days", *dayPolicy.Schedule.Unit)
		assert.Equal(t, int64(7), *dayPolicy.Schedule.DaySchedule.Frequency)

		monthPolicy, err := GenerateFullSchedule(types.FullSchedule_Unit_Months, &types.FullBackup{
			MonthSchedule: &types.MonthSchedule{OnDate: 15},
		})
		assert.NoError(t, err)
		assert.Equal(t, "Months", *monthPolicy.Schedule.Unit)
		assert.Equal(t, int64(15), *monthPolicy.Schedule.MonthSchedule.DayOfMonth)

		weekPolicy, err := GenerateFullSchedule(types.FullSchedule_Unit_Weeks, &types.FullBackup{
			WeekSchedule: &types.WeekSchedule{OnDayOfWeek: []types.DayOfWeek{types.WeekDay_Sunday}},
		})
		assert.NoError(t, err)
		assert.Equal(t, "Weeks", *weekPolicy.Schedule.Unit)
		assert.Equal(t, []string{"Sunday"}, weekPolicy.Schedule.WeekSchedule.DayOfWeek)

		yearDay := types.YearDay_Last
		yearPolicy, err := GenerateFullSchedule(types.FullSchedule_Unit_Years, &types.FullBackup{
			YearSchedule: &types.YearSchedule{OnDayOfYear: &yearDay},
		})
		assert.NoError(t, err)
		assert.Equal(t, "Years", *yearPolicy.Schedule.Unit)
		assert.Equal(t, "Last", *yearPolicy.Schedule.YearSchedule.DayOfYear)
	})

	t.Run("returns error when unit specific schedule is missing", func(t *testing.T) {
		result, err := GenerateFullSchedule(types.FullSchedule_Unit_Days, &types.FullBackup{})

		assert.Nil(t, result)
		assert.EqualError(t, err, "Day schedule needs to be set for Unit set to Days")
	})

	t.Run("returns unit-specific errors for missing schedule payloads", func(t *testing.T) {
		monthResult, monthErr := GenerateFullSchedule(types.FullSchedule_Unit_Months, &types.FullBackup{})
		assert.Nil(t, monthResult)
		assert.EqualError(t, monthErr, "Hour schedule needs to be set for Unit set to Hours")

		weekResult, weekErr := GenerateFullSchedule(types.FullSchedule_Unit_Weeks, &types.FullBackup{})
		assert.Nil(t, weekResult)
		assert.EqualError(t, weekErr, "Minute schedule needs to be set for Unit set to Minute")

		yearResult, yearErr := GenerateFullSchedule(types.FullSchedule_Unit_Years, &types.FullBackup{})
		assert.Nil(t, yearResult)
		assert.EqualError(t, yearErr, "Month schedule needs to be set for Unit set to Month")
	})

	t.Run("returns error for unsupported unit", func(t *testing.T) {
		result, err := GenerateFullSchedule(types.FullSchedule_Unit("Invalid"), &types.FullBackup{})

		assert.Nil(t, result)
		assert.EqualError(t, err, "Invalid incremental schedule unit provided")
	})
}

func TestBuildRetention(t *testing.T) {
	t.Run("returns error when required fields are missing", func(t *testing.T) {
		result, err := buildRetention(&types.DataRetention{})

		assert.Nil(t, result)
		assert.EqualError(t, err, " `Unit` and `RetainFor` are required parameters")
	})

	t.Run("builds retention without data lock config", func(t *testing.T) {
		unit := types.Retention_Unit_Weeks
		result, err := buildRetention(&types.DataRetention{
			Unit:      &unit,
			RetainFor: 8,
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Weeks", *result.Unit)
		assert.Equal(t, int64(8), *result.Duration)
		assert.Nil(t, result.DataLockConfig)
	})

	t.Run("builds retention with data lock config", func(t *testing.T) {
		retentionUnit := types.Retention_Unit_Years
		mode := types.DataLockConfig_Mode_Compliance
		lockUnit := types.DataLockConfig_Unit_Months
		result, err := buildRetention(&types.DataRetention{
			Unit:      &retentionUnit,
			RetainFor: 2,
			DataLockConfig: &types.DataLockConfig{
				Mode:                       &mode,
				Unit:                       &lockUnit,
				Duration:                   6,
				EnableWormOnExternalTarget: true,
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.DataLockConfig)
		assert.Equal(t, "Compliance", *result.DataLockConfig.Mode)
		assert.Equal(t, "Months", *result.DataLockConfig.Unit)
		assert.Equal(t, int64(6), *result.DataLockConfig.Duration)
		assert.Equal(t, true, *result.DataLockConfig.EnableWormOnExternalTarget)
	})

	t.Run("returns error when data lock config is invalid", func(t *testing.T) {
		retentionUnit := types.Retention_Unit_Years
		result, err := buildRetention(&types.DataRetention{
			Unit:           &retentionUnit,
			RetainFor:      2,
			DataLockConfig: &types.DataLockConfig{},
		})

		assert.Nil(t, result)
		assert.EqualError(t, err, "`Mode`, `Unit` and `Duration` are required parameters")
	})
}

func TestBuildDataLockConfig(t *testing.T) {
	t.Run("returns error when required fields are missing", func(t *testing.T) {
		result, err := buildDataLockConfig(&types.DataLockConfig{})

		assert.Nil(t, result)
		assert.EqualError(t, err, "`Mode`, `Unit` and `Duration` are required parameters")
	})

	t.Run("builds data lock config", func(t *testing.T) {
		mode := types.DataLockConfig_Mode_Administrative
		unit := types.DataLockConfig_Unit_Days
		result, err := buildDataLockConfig(&types.DataLockConfig{
			Mode:                       &mode,
			Unit:                       &unit,
			Duration:                   30,
			EnableWormOnExternalTarget: false,
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Administrative", *result.Mode)
		assert.Equal(t, "Days", *result.Unit)
		assert.Equal(t, int64(30), *result.Duration)
		assert.Equal(t, false, *result.EnableWormOnExternalTarget)
	})
}

// Made with Bob
