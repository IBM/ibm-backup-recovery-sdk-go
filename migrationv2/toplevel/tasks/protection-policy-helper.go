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
	"fmt"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

func buildRetryOptions(retryOption types.RetryOption) *backuprecoveryv1.RetryOptions {
	if retryOption.NumberOfRetry == 0 {
		// default
		retries := types.DefaultPolicyRetries
		interval := types.DefaultPolicyInterval
		return &backuprecoveryv1.RetryOptions{
			Retries:           &retries,
			RetryIntervalMins: &interval,
		}
	}

	return &backuprecoveryv1.RetryOptions{
		Retries:           core.Int64Ptr(retryOption.NumberOfRetry),
		RetryIntervalMins: core.Int64Ptr(retryOption.WaitInCaseOfFailure),
	}
}

func buildIncrementalSchedule(incrementalBackup *types.IncrementalBackup) (*backuprecoveryv1.IncrementalBackupPolicy, error) {

	if incrementalBackup.Unit == nil {
		return nil, fmt.Errorf("`Unit` is a required field for incrementalBackup")
	}
	return GenerateIncrementalSchedule(*incrementalBackup.Unit, incrementalBackup)
}

func GenerateIncrementalSchedule(unit types.IncrementalBackup_Unit, incrementalBackup *types.IncrementalBackup) (*backuprecoveryv1.IncrementalBackupPolicy, error) {

	switch unit {
	// build day specific schedule when unit set to "Days"
	case types.IncrementalBackup_Unit_Days:
		if incrementalBackup.DaySchedule == nil {
			return nil, fmt.Errorf("Day schedule needs to be set for Unit set to Days")
		}
		return GenerateIncrementalDaySchedule(string(unit), incrementalBackup.DaySchedule), nil
	// build hour specific schedule when unit set to "Hours"
	case types.IncrementalBackup_Unit_Hours:
		if incrementalBackup.HourSchedule == nil {
			return nil, fmt.Errorf("Hour schedule needs to be set for Unit set to Hours")
		}
		return GenerateIncrementalHourSchedule(string(unit), incrementalBackup.HourSchedule), nil
	// build minute specific schedule when unit set to "Hours"
	case types.IncrementalBackup_Unit_Minutes:
		if incrementalBackup.MinuteSchedule == nil {
			return nil, fmt.Errorf("Minute schedule needs to be set for Unit set to Minute")
		}
		return GenerateIncrementalMinuteSchedule(string(unit), incrementalBackup.MinuteSchedule), nil
	case types.IncrementalBackup_Unit_Months:
		if incrementalBackup.MonthSchedule == nil {
			return nil, fmt.Errorf("Month schedule needs to be set for Unit set to Month")
		}
		return GenerateIncrementalMonthSchedule(string(unit), incrementalBackup.MonthSchedule), nil
	case types.IncrementalBackup_Unit_Weeks:
		if incrementalBackup.WeekSchedule == nil {
			return nil, fmt.Errorf("Week schedule needs to be set for Unit set to Week")
		}
		return GenerateIncrementalWeekSchedule(string(unit), incrementalBackup.WeekSchedule), nil
	case types.IncrementalBackup_Unit_Years:
		if incrementalBackup.YearSchedule == nil {
			return nil, fmt.Errorf("Year schedule needs to be set for Unit set to Year")
		}
		return GenerateIncrementalYearSchedule(string(unit), incrementalBackup.YearSchedule), nil
	default:
		return nil, fmt.Errorf("Invalid incremental schedule unit provided")
	}

}

func GenerateIncrementalDaySchedule(unit string, daySchedule *types.UnitDaySchedule) *backuprecoveryv1.IncrementalBackupPolicy {
	return &backuprecoveryv1.IncrementalBackupPolicy{
		Schedule: &backuprecoveryv1.IncrementalSchedule{
			Unit: core.StringPtr(unit),
			DaySchedule: &backuprecoveryv1.DaySchedule{
				Frequency: &daySchedule.Every,
			},
		}}
}

func GenerateIncrementalHourSchedule(unit string, hourSchedule *types.UnitDaySchedule) *backuprecoveryv1.IncrementalBackupPolicy {
	return &backuprecoveryv1.IncrementalBackupPolicy{
		Schedule: &backuprecoveryv1.IncrementalSchedule{
			Unit: core.StringPtr(unit),
			HourSchedule: &backuprecoveryv1.HourSchedule{
				Frequency: &hourSchedule.Every,
			},
		}}
}

func GenerateIncrementalMinuteSchedule(unit string, minuteSchedule *types.UnitDaySchedule) *backuprecoveryv1.IncrementalBackupPolicy {
	return &backuprecoveryv1.IncrementalBackupPolicy{
		Schedule: &backuprecoveryv1.IncrementalSchedule{
			Unit: core.StringPtr(unit),
			MinuteSchedule: &backuprecoveryv1.MinuteSchedule{
				Frequency: &minuteSchedule.Every,
			},
		}}
}

func GenerateIncrementalMonthSchedule(unit string, monthSchedule *types.MonthSchedule) *backuprecoveryv1.IncrementalBackupPolicy {
	return &backuprecoveryv1.IncrementalBackupPolicy{
		Schedule: &backuprecoveryv1.IncrementalSchedule{
			Unit: core.StringPtr(unit),
			MonthSchedule: &backuprecoveryv1.MonthSchedule{
				DayOfMonth: &monthSchedule.OnDate,
			},
		}}
}

func GenerateIncrementalWeekSchedule(unit string, weekSchedule *types.WeekSchedule) *backuprecoveryv1.IncrementalBackupPolicy {
	dayOfWeek := make([]string, 0)
	for _, day := range weekSchedule.OnDayOfWeek {
		dayOfWeek = append(dayOfWeek, string(day))
	}
	return &backuprecoveryv1.IncrementalBackupPolicy{
		Schedule: &backuprecoveryv1.IncrementalSchedule{
			Unit: core.StringPtr(unit),
			WeekSchedule: &backuprecoveryv1.WeekSchedule{
				DayOfWeek: dayOfWeek,
			},
		}}
}

func GenerateIncrementalYearSchedule(unit string, yearSchedule *types.YearSchedule) *backuprecoveryv1.IncrementalBackupPolicy {
	return &backuprecoveryv1.IncrementalBackupPolicy{
		Schedule: &backuprecoveryv1.IncrementalSchedule{
			Unit: core.StringPtr(unit),
			YearSchedule: &backuprecoveryv1.YearSchedule{
				DayOfYear: (*string)(yearSchedule.OnDayOfYear),
			},
		}}
}

func buildFullSchedule(fullBackups []types.FullBackup) ([]backuprecoveryv1.FullScheduleAndRetention, error) {
	if len(fullBackups) == 0 {
		return nil, nil
	}
	fullScheduleAndRetentions := make([]backuprecoveryv1.FullScheduleAndRetention, 0)
	for _, fullBackup := range fullBackups {
		if fullBackup.Unit == nil || fullBackup.Retention == nil {
			return nil, fmt.Errorf("`Unit` and `Retention` are required fields for a FullBackup")
		}

		fullScheduleAndRetention, err := GenerateFullSchedule(*fullBackup.Unit, &fullBackup)
		if err != nil {
			return nil, err
		}

		retention, err := buildRetention(fullBackup.Retention)
		if err != nil {
			return nil, err
		}
		fullScheduleAndRetention.Retention = retention

		fullScheduleAndRetentions = append(fullScheduleAndRetentions, *fullScheduleAndRetention)
	}
	return fullScheduleAndRetentions, nil

}

func GenerateFullSchedule(unit types.FullSchedule_Unit, fullBackup *types.FullBackup) (*backuprecoveryv1.FullScheduleAndRetention, error) {

	switch unit {
	case types.FullSchedule_Unit_Days:
		if fullBackup.DaySchedule == nil {
			return nil, fmt.Errorf("Day schedule needs to be set for Unit set to Days")
		}
		return GenerateFullDaySchedule(string(unit), fullBackup.DaySchedule), nil
	case types.FullSchedule_Unit_Months:
		if fullBackup.MonthSchedule == nil {
			return nil, fmt.Errorf("Hour schedule needs to be set for Unit set to Hours")
		}
		return GenerateFullMonthSchedule(string(unit), fullBackup.MonthSchedule), nil
	case types.FullSchedule_Unit_Weeks:
		if fullBackup.WeekSchedule == nil {
			return nil, fmt.Errorf("Minute schedule needs to be set for Unit set to Minute")
		}
		return GenerateFullWeekSchedule(string(unit), fullBackup.WeekSchedule), nil
	case types.FullSchedule_Unit_Years:
		if fullBackup.YearSchedule == nil {
			return nil, fmt.Errorf("Month schedule needs to be set for Unit set to Month")
		}
		return GenerateFullYearSchedule(string(unit), fullBackup.YearSchedule), nil
	default:
		return nil, fmt.Errorf("Invalid incremental schedule unit provided")
	}

}

func GenerateFullDaySchedule(unit string, daySchedule *types.UnitDaySchedule) *backuprecoveryv1.FullScheduleAndRetention {
	return &backuprecoveryv1.FullScheduleAndRetention{
		Schedule: &backuprecoveryv1.FullSchedule{
			Unit: core.StringPtr(unit),
			DaySchedule: &backuprecoveryv1.DaySchedule{
				Frequency: &daySchedule.Every,
			},
		}}
}

func GenerateFullMonthSchedule(unit string, monthSchedule *types.MonthSchedule) *backuprecoveryv1.FullScheduleAndRetention {
	return &backuprecoveryv1.FullScheduleAndRetention{
		Schedule: &backuprecoveryv1.FullSchedule{
			Unit: core.StringPtr(unit),
			MonthSchedule: &backuprecoveryv1.MonthSchedule{
				DayOfMonth: &monthSchedule.OnDate,
			},
		}}
}

func GenerateFullWeekSchedule(unit string, weekSchedule *types.WeekSchedule) *backuprecoveryv1.FullScheduleAndRetention {
	dayOfWeek := make([]string, 0)
	for _, day := range weekSchedule.OnDayOfWeek {
		dayOfWeek = append(dayOfWeek, string(day))
	}
	return &backuprecoveryv1.FullScheduleAndRetention{
		Schedule: &backuprecoveryv1.FullSchedule{
			Unit: core.StringPtr(unit),
			WeekSchedule: &backuprecoveryv1.WeekSchedule{
				DayOfWeek: dayOfWeek,
			},
		}}
}

func GenerateFullYearSchedule(unit string, yearSchedule *types.YearSchedule) *backuprecoveryv1.FullScheduleAndRetention {
	return &backuprecoveryv1.FullScheduleAndRetention{
		Schedule: &backuprecoveryv1.FullSchedule{
			Unit: core.StringPtr(unit),
			YearSchedule: &backuprecoveryv1.YearSchedule{
				DayOfYear: (*string)(yearSchedule.OnDayOfYear),
			},
		}}
}

func buildRetention(retention *types.DataRetention) (*backuprecoveryv1.Retention, error) {

	if retention.Unit == nil || retention.RetainFor == 0 {
		return nil, fmt.Errorf(" `Unit` and `RetainFor` are required parameters")
	}

	retentionResult := &backuprecoveryv1.Retention{}
	if retention.DataLockConfig != nil {
		datalockConfig, err := buildDataLockConfig(retention.DataLockConfig)
		if err != nil {
			return nil, err
		}
		retentionResult.DataLockConfig = datalockConfig
	}
	retentionResult.Unit = (*string)(retention.Unit)
	retentionResult.Duration = &retention.RetainFor

	return retentionResult, nil

}

func buildDataLockConfig(dataLockConfig *types.DataLockConfig) (*backuprecoveryv1.DataLockConfig, error) {

	if dataLockConfig.Mode == nil || dataLockConfig.Unit == nil || dataLockConfig.Duration == 0 {
		return nil, fmt.Errorf("`Mode`, `Unit` and `Duration` are required parameters")
	}

	return &backuprecoveryv1.DataLockConfig{
		Duration:                   &dataLockConfig.Duration,
		EnableWormOnExternalTarget: &dataLockConfig.EnableWormOnExternalTarget,
		Mode:                       (*string)(dataLockConfig.Mode),
		Unit:                       (*string)(dataLockConfig.Unit),
	}, nil

}
