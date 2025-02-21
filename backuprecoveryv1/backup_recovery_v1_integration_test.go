//go:build integration

/**
 * (C) Copyright IBM Corp. 2025.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package backuprecoveryv1_test

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

/**
 * This file contains an integration test for the backuprecoveryv1 package.
 *
 * Notes:
 *
 * The integration test will automatically skip tests if the required config file is not available.
 */

var _ = Describe(`BackupRecoveryV1 Integration Tests`, func() {
	const externalConfigFile = "../backup_recovery_v1.env"

	var (
		err                   error
		backupRecoveryService *backuprecoveryv1.BackupRecoveryV1
		serviceURL            string
		config                map[string]string

		// Variables to hold link values
		connectionIdLink       string
		policyIdLink           string
		protectionGroupIdLink  string
		protectionSourceIdLink int64
		recoveryIdLink         string
	)

	var shouldSkipTest = func() {
		Skip("External configuration is not available, skipping tests...")
	}

	Describe(`External configuration`, func() {
		It("Successfully load the configuration", func() {
			_, err = os.Stat(externalConfigFile)
			if err != nil {
				Skip("External configuration file not found, skipping tests: " + err.Error())
			}

			os.Setenv("IBM_CREDENTIALS_FILE", externalConfigFile)
			config, err = core.GetServiceProperties(backuprecoveryv1.DefaultServiceName)
			if err != nil {
				Skip("Error loading service properties, skipping tests: " + err.Error())
			}
			serviceURL = config["URL"]
			if serviceURL == "" {
				Skip("Unable to load service URL configuration property, skipping tests")
			}

			fmt.Fprintf(GinkgoWriter, "Service URL: %v\n", serviceURL)
			shouldSkipTest = func() {}
		})
	})

	Describe(`Client initialization`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It("Successfully construct the service client instance", func() {
			backupRecoveryServiceOptions := &backuprecoveryv1.BackupRecoveryV1Options{}

			backupRecoveryService, err = backuprecoveryv1.NewBackupRecoveryV1UsingExternalConfig(backupRecoveryServiceOptions)
			Expect(err).To(BeNil())
			Expect(backupRecoveryService).ToNot(BeNil())
			Expect(backupRecoveryService.Service.Options.URL).To(Equal(serviceURL))

			core.SetLogger(core.NewLogger(core.LevelDebug, log.New(GinkgoWriter, "", log.LstdFlags), log.New(GinkgoWriter, "", log.LstdFlags)))
			backupRecoveryService.EnableRetries(4, 30*time.Second)
		})
	})

	Describe(`CreateDataSourceConnection - Create a data-source connection`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`CreateDataSourceConnection(createDataSourceConnectionOptions *CreateDataSourceConnectionOptions)`, func() {
			createDataSourceConnectionOptions := &backuprecoveryv1.CreateDataSourceConnectionOptions{
				ConnectionName: core.StringPtr("data-source-connection"),
				XIBMTenantID:   core.StringPtr("tenantId"),
			}

			dataSourceConnection, response, err := backupRecoveryService.CreateDataSourceConnection(createDataSourceConnectionOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(201))
			Expect(dataSourceConnection).ToNot(BeNil())

			connectionIdLink = *dataSourceConnection.ConnectionID
			fmt.Fprintf(GinkgoWriter, "Saved connectionIdLink value: %v\n", connectionIdLink)
		})
	})

	Describe(`CreateProtectionPolicy - Create a Protection Policy`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`CreateProtectionPolicy(createProtectionPolicyOptions *CreateProtectionPolicyOptions)`, func() {
			minuteScheduleModel := &backuprecoveryv1.MinuteSchedule{
				Frequency: core.Int64Ptr(int64(1)),
			}

			hourScheduleModel := &backuprecoveryv1.HourSchedule{
				Frequency: core.Int64Ptr(int64(1)),
			}

			dayScheduleModel := &backuprecoveryv1.DaySchedule{
				Frequency: core.Int64Ptr(int64(1)),
			}

			weekScheduleModel := &backuprecoveryv1.WeekSchedule{
				DayOfWeek: []string{"Sunday"},
			}

			monthScheduleModel := &backuprecoveryv1.MonthSchedule{
				DayOfWeek:   []string{"Sunday"},
				WeekOfMonth: core.StringPtr("First"),
				DayOfMonth:  core.Int64Ptr(int64(10)),
			}

			yearScheduleModel := &backuprecoveryv1.YearSchedule{
				DayOfYear: core.StringPtr("First"),
			}

			incrementalScheduleModel := &backuprecoveryv1.IncrementalSchedule{
				Unit:           core.StringPtr("Minutes"),
				MinuteSchedule: minuteScheduleModel,
				HourSchedule:   hourScheduleModel,
				DaySchedule:    dayScheduleModel,
				WeekSchedule:   weekScheduleModel,
				MonthSchedule:  monthScheduleModel,
				YearSchedule:   yearScheduleModel,
			}

			incrementalBackupPolicyModel := &backuprecoveryv1.IncrementalBackupPolicy{
				Schedule: incrementalScheduleModel,
			}

			fullScheduleModel := &backuprecoveryv1.FullSchedule{
				Unit:          core.StringPtr("Days"),
				DaySchedule:   dayScheduleModel,
				WeekSchedule:  weekScheduleModel,
				MonthSchedule: monthScheduleModel,
				YearSchedule:  yearScheduleModel,
			}

			fullBackupPolicyModel := &backuprecoveryv1.FullBackupPolicy{
				Schedule: fullScheduleModel,
			}

			dataLockConfigModel := &backuprecoveryv1.DataLockConfig{
				Mode:                       core.StringPtr("Compliance"),
				Unit:                       core.StringPtr("Days"),
				Duration:                   core.Int64Ptr(int64(1)),
				EnableWormOnExternalTarget: core.BoolPtr(true),
			}

			retentionModel := &backuprecoveryv1.Retention{
				Unit:           core.StringPtr("Days"),
				Duration:       core.Int64Ptr(int64(1)),
				DataLockConfig: dataLockConfigModel,
			}

			fullScheduleAndRetentionModel := &backuprecoveryv1.FullScheduleAndRetention{
				Schedule:  fullScheduleModel,
				Retention: retentionModel,
			}

			awsTierModel := &backuprecoveryv1.AWSTier{
				MoveAfterUnit: core.StringPtr("Days"),
				MoveAfter:     core.Int64Ptr(int64(26)),
				TierType:      core.StringPtr("kAmazonS3Standard"),
			}

			awsTiersModel := &backuprecoveryv1.AWSTiers{
				Tiers: []backuprecoveryv1.AWSTier{*awsTierModel},
			}

			azureTierModel := &backuprecoveryv1.AzureTier{
				MoveAfterUnit: core.StringPtr("Days"),
				MoveAfter:     core.Int64Ptr(int64(26)),
				TierType:      core.StringPtr("kAzureTierHot"),
			}

			azureTiersModel := &backuprecoveryv1.AzureTiers{
				Tiers: []backuprecoveryv1.AzureTier{*azureTierModel},
			}

			googleTierModel := &backuprecoveryv1.GoogleTier{
				MoveAfterUnit: core.StringPtr("Days"),
				MoveAfter:     core.Int64Ptr(int64(26)),
				TierType:      core.StringPtr("kGoogleStandard"),
			}

			googleTiersModel := &backuprecoveryv1.GoogleTiers{
				Tiers: []backuprecoveryv1.GoogleTier{*googleTierModel},
			}

			oracleTierModel := &backuprecoveryv1.OracleTier{
				MoveAfterUnit: core.StringPtr("Days"),
				MoveAfter:     core.Int64Ptr(int64(26)),
				TierType:      core.StringPtr("kOracleTierStandard"),
			}

			oracleTiersModel := &backuprecoveryv1.OracleTiers{
				Tiers: []backuprecoveryv1.OracleTier{*oracleTierModel},
			}

			tierLevelSettingsModel := &backuprecoveryv1.TierLevelSettings{
				AwsTiering:    awsTiersModel,
				AzureTiering:  azureTiersModel,
				CloudPlatform: core.StringPtr("AWS"),
				GoogleTiering: googleTiersModel,
				OracleTiering: oracleTiersModel,
			}

			primaryArchivalTargetModel := &backuprecoveryv1.PrimaryArchivalTarget{
				TargetID:     core.Int64Ptr(int64(26)),
				TierSettings: tierLevelSettingsModel,
			}

			primaryBackupTargetModel := &backuprecoveryv1.PrimaryBackupTarget{
				TargetType:             core.StringPtr("Local"),
				ArchivalTargetSettings: primaryArchivalTargetModel,
				UseDefaultBackupTarget: core.BoolPtr(true),
			}

			regularBackupPolicyModel := &backuprecoveryv1.RegularBackupPolicy{
				Incremental:         incrementalBackupPolicyModel,
				Full:                fullBackupPolicyModel,
				FullBackups:         []backuprecoveryv1.FullScheduleAndRetention{*fullScheduleAndRetentionModel},
				Retention:           retentionModel,
				PrimaryBackupTarget: primaryBackupTargetModel,
			}

			logScheduleModel := &backuprecoveryv1.LogSchedule{
				Unit:           core.StringPtr("Minutes"),
				MinuteSchedule: minuteScheduleModel,
				HourSchedule:   hourScheduleModel,
			}

			logBackupPolicyModel := &backuprecoveryv1.LogBackupPolicy{
				Schedule:  logScheduleModel,
				Retention: retentionModel,
			}

			bmrScheduleModel := &backuprecoveryv1.BmrSchedule{
				Unit:          core.StringPtr("Days"),
				DaySchedule:   dayScheduleModel,
				WeekSchedule:  weekScheduleModel,
				MonthSchedule: monthScheduleModel,
				YearSchedule:  yearScheduleModel,
			}

			bmrBackupPolicyModel := &backuprecoveryv1.BmrBackupPolicy{
				Schedule:  bmrScheduleModel,
				Retention: retentionModel,
			}

			cdpRetentionModel := &backuprecoveryv1.CdpRetention{
				Unit:           core.StringPtr("Minutes"),
				Duration:       core.Int64Ptr(int64(1)),
				DataLockConfig: dataLockConfigModel,
			}

			cdpBackupPolicyModel := &backuprecoveryv1.CdpBackupPolicy{
				Retention: cdpRetentionModel,
			}

			storageArraySnapshotScheduleModel := &backuprecoveryv1.StorageArraySnapshotSchedule{
				Unit:           core.StringPtr("Minutes"),
				MinuteSchedule: minuteScheduleModel,
				HourSchedule:   hourScheduleModel,
				DaySchedule:    dayScheduleModel,
				WeekSchedule:   weekScheduleModel,
				MonthSchedule:  monthScheduleModel,
				YearSchedule:   yearScheduleModel,
			}

			storageArraySnapshotBackupPolicyModel := &backuprecoveryv1.StorageArraySnapshotBackupPolicy{
				Schedule:  storageArraySnapshotScheduleModel,
				Retention: retentionModel,
			}

			cancellationTimeoutParamsModel := &backuprecoveryv1.CancellationTimeoutParams{
				TimeoutMins: core.Int64Ptr(int64(26)),
				BackupType:  core.StringPtr("kRegular"),
			}

			backupPolicyModel := &backuprecoveryv1.BackupPolicy{
				Regular:              regularBackupPolicyModel,
				Log:                  logBackupPolicyModel,
				Bmr:                  bmrBackupPolicyModel,
				Cdp:                  cdpBackupPolicyModel,
				StorageArraySnapshot: storageArraySnapshotBackupPolicyModel,
				RunTimeouts:          []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
			}

			timeOfDayModel := &backuprecoveryv1.TimeOfDay{
				Hour:     core.Int64Ptr(int64(1)),
				Minute:   core.Int64Ptr(int64(15)),
				TimeZone: core.StringPtr("America/Los_Angeles"),
			}

			blackoutWindowModel := &backuprecoveryv1.BlackoutWindow{
				Day:       core.StringPtr("Sunday"),
				StartTime: timeOfDayModel,
				EndTime:   timeOfDayModel,
				ConfigID:  core.StringPtr("Config-Id"),
			}

			extendedRetentionScheduleModel := &backuprecoveryv1.ExtendedRetentionSchedule{
				Unit:      core.StringPtr("Runs"),
				Frequency: core.Int64Ptr(int64(3)),
			}

			extendedRetentionPolicyModel := &backuprecoveryv1.ExtendedRetentionPolicy{
				Schedule:  extendedRetentionScheduleModel,
				Retention: retentionModel,
				RunType:   core.StringPtr("Regular"),
				ConfigID:  core.StringPtr("Config-Id"),
			}

			targetScheduleModel := &backuprecoveryv1.TargetSchedule{
				Unit:      core.StringPtr("Runs"),
				Frequency: core.Int64Ptr(int64(3)),
			}

			logRetentionModel := &backuprecoveryv1.LogRetention{
				Unit:           core.StringPtr("Days"),
				Duration:       core.Int64Ptr(int64(0)),
				DataLockConfig: dataLockConfigModel,
			}

			awsTargetConfigModel := &backuprecoveryv1.AWSTargetConfig{
				Region:   core.Int64Ptr(int64(26)),
				SourceID: core.Int64Ptr(int64(26)),
			}

			azureTargetConfigModel := &backuprecoveryv1.AzureTargetConfig{
				ResourceGroup: core.Int64Ptr(int64(26)),
				SourceID:      core.Int64Ptr(int64(26)),
			}

			remoteTargetConfigModel := &backuprecoveryv1.RemoteTargetConfig{
				ClusterID: core.Int64Ptr(int64(26)),
			}

			replicationTargetConfigurationModel := &backuprecoveryv1.ReplicationTargetConfiguration{
				Schedule:           targetScheduleModel,
				Retention:          retentionModel,
				CopyOnRunSuccess:   core.BoolPtr(true),
				ConfigID:           core.StringPtr("Config-Id"),
				BackupRunType:      core.StringPtr("Regular"),
				RunTimeouts:        []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				LogRetention:       logRetentionModel,
				AwsTargetConfig:    awsTargetConfigModel,
				AzureTargetConfig:  azureTargetConfigModel,
				TargetType:         core.StringPtr("RemoteCluster"),
				RemoteTargetConfig: remoteTargetConfigModel,
			}

			archivalTargetConfigurationModel := &backuprecoveryv1.ArchivalTargetConfiguration{
				Schedule:          targetScheduleModel,
				Retention:         retentionModel,
				CopyOnRunSuccess:  core.BoolPtr(true),
				ConfigID:          core.StringPtr("Config-Id"),
				BackupRunType:     core.StringPtr("Regular"),
				RunTimeouts:       []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				LogRetention:      logRetentionModel,
				TargetID:          core.Int64Ptr(int64(5)),
				TierSettings:      tierLevelSettingsModel,
				ExtendedRetention: []backuprecoveryv1.ExtendedRetentionPolicy{*extendedRetentionPolicyModel},
			}

			customTagParamsModel := &backuprecoveryv1.CustomTagParams{
				Key:   core.StringPtr("custom-tag-key"),
				Value: core.StringPtr("custom-tag-value"),
			}

			awsCloudSpinParamsModel := &backuprecoveryv1.AwsCloudSpinParams{
				CustomTagList: []backuprecoveryv1.CustomTagParams{*customTagParamsModel},
				Region:        core.Int64Ptr(int64(3)),
				SubnetID:      core.Int64Ptr(int64(26)),
				VpcID:         core.Int64Ptr(int64(26)),
			}

			azureCloudSpinParamsModel := &backuprecoveryv1.AzureCloudSpinParams{
				AvailabilitySetID:        core.Int64Ptr(int64(26)),
				NetworkResourceGroupID:   core.Int64Ptr(int64(26)),
				ResourceGroupID:          core.Int64Ptr(int64(26)),
				StorageAccountID:         core.Int64Ptr(int64(26)),
				StorageContainerID:       core.Int64Ptr(int64(26)),
				StorageResourceGroupID:   core.Int64Ptr(int64(26)),
				TempVmResourceGroupID:    core.Int64Ptr(int64(26)),
				TempVmStorageAccountID:   core.Int64Ptr(int64(26)),
				TempVmStorageContainerID: core.Int64Ptr(int64(26)),
				TempVmSubnetID:           core.Int64Ptr(int64(26)),
				TempVmVirtualNetworkID:   core.Int64Ptr(int64(26)),
			}

			cloudSpinTargetModel := &backuprecoveryv1.CloudSpinTarget{
				AwsParams:   awsCloudSpinParamsModel,
				AzureParams: azureCloudSpinParamsModel,
				ID:          core.Int64Ptr(int64(2)),
			}

			cloudSpinTargetConfigurationModel := &backuprecoveryv1.CloudSpinTargetConfiguration{
				Schedule:         targetScheduleModel,
				Retention:        retentionModel,
				CopyOnRunSuccess: core.BoolPtr(true),
				ConfigID:         core.StringPtr("Config-Id"),
				BackupRunType:    core.StringPtr("Regular"),
				RunTimeouts:      []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				LogRetention:     logRetentionModel,
				Target:           cloudSpinTargetModel,
			}

			onpremDeployParamsModel := &backuprecoveryv1.OnpremDeployParams{
				ID: core.Int64Ptr(int64(4)),
			}

			onpremDeployTargetConfigurationModel := &backuprecoveryv1.OnpremDeployTargetConfiguration{
				Schedule:         targetScheduleModel,
				Retention:        retentionModel,
				CopyOnRunSuccess: core.BoolPtr(true),
				ConfigID:         core.StringPtr("Config-Id"),
				BackupRunType:    core.StringPtr("Regular"),
				RunTimeouts:      []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				LogRetention:     logRetentionModel,
				Params:           onpremDeployParamsModel,
			}

			rpaasTargetConfigurationModel := &backuprecoveryv1.RpaasTargetConfiguration{
				Schedule:         targetScheduleModel,
				Retention:        retentionModel,
				CopyOnRunSuccess: core.BoolPtr(true),
				ConfigID:         core.StringPtr("Config-Id"),
				BackupRunType:    core.StringPtr("Regular"),
				RunTimeouts:      []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				LogRetention:     logRetentionModel,
				TargetID:         core.Int64Ptr(int64(5)),
				TargetType:       core.StringPtr("Tape"),
			}

			targetsConfigurationModel := &backuprecoveryv1.TargetsConfiguration{
				ReplicationTargets:  []backuprecoveryv1.ReplicationTargetConfiguration{*replicationTargetConfigurationModel},
				ArchivalTargets:     []backuprecoveryv1.ArchivalTargetConfiguration{*archivalTargetConfigurationModel},
				CloudSpinTargets:    []backuprecoveryv1.CloudSpinTargetConfiguration{*cloudSpinTargetConfigurationModel},
				OnpremDeployTargets: []backuprecoveryv1.OnpremDeployTargetConfiguration{*onpremDeployTargetConfigurationModel},
				RpaasTargets:        []backuprecoveryv1.RpaasTargetConfiguration{*rpaasTargetConfigurationModel},
			}

			cascadedTargetConfigurationModel := &backuprecoveryv1.CascadedTargetConfiguration{
				SourceClusterID: core.Int64Ptr(int64(26)),
				RemoteTargets:   targetsConfigurationModel,
			}

			retryOptionsModel := &backuprecoveryv1.RetryOptions{
				Retries:           core.Int64Ptr(int64(0)),
				RetryIntervalMins: core.Int64Ptr(int64(1)),
			}

			createProtectionPolicyOptions := &backuprecoveryv1.CreateProtectionPolicyOptions{
				XIBMTenantID:              core.StringPtr("tenantId"),
				Name:                      core.StringPtr("create-protection-policy"),
				BackupPolicy:              backupPolicyModel,
				Description:               core.StringPtr("Protection Policy"),
				BlackoutWindow:            []backuprecoveryv1.BlackoutWindow{*blackoutWindowModel},
				ExtendedRetention:         []backuprecoveryv1.ExtendedRetentionPolicy{*extendedRetentionPolicyModel},
				RemoteTargetPolicy:        targetsConfigurationModel,
				CascadedTargetsConfig:     []backuprecoveryv1.CascadedTargetConfiguration{*cascadedTargetConfigurationModel},
				RetryOptions:              retryOptionsModel,
				DataLock:                  core.StringPtr("Compliance"),
				Version:                   core.Int64Ptr(int64(38)),
				IsCBSEnabled:              core.BoolPtr(true),
				LastModificationTimeUsecs: core.Int64Ptr(int64(26)),
				TemplateID:                core.StringPtr("protection-policy-template"),
			}

			protectionPolicyResponse, response, err := backupRecoveryService.CreateProtectionPolicy(createProtectionPolicyOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(201))
			Expect(protectionPolicyResponse).ToNot(BeNil())

			policyIdLink = *protectionPolicyResponse.ID
			fmt.Fprintf(GinkgoWriter, "Saved policyIdLink value: %v\n", policyIdLink)
		})
	})

	Describe(`CreateProtectionGroup - Create a Protection Group`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`CreateProtectionGroup(createProtectionGroupOptions *CreateProtectionGroupOptions)`, func() {
			timeOfDayModel := &backuprecoveryv1.TimeOfDay{
				Hour:     core.Int64Ptr(int64(0)),
				Minute:   core.Int64Ptr(int64(0)),
				TimeZone: core.StringPtr("America/Los_Angeles"),
			}

			alertTargetModel := &backuprecoveryv1.AlertTarget{
				EmailAddress:  core.StringPtr("alert1@domain.com"),
				Language:      core.StringPtr("en-us"),
				RecipientType: core.StringPtr("kTo"),
			}

			protectionGroupAlertingPolicyModel := &backuprecoveryv1.ProtectionGroupAlertingPolicy{
				BackupRunStatus:              []string{"kSuccess"},
				AlertTargets:                 []backuprecoveryv1.AlertTarget{*alertTargetModel},
				RaiseObjectLevelFailureAlert: core.BoolPtr(true),
				RaiseObjectLevelFailureAlertAfterLastAttempt: core.BoolPtr(true),
				RaiseObjectLevelFailureAlertAfterEachAttempt: core.BoolPtr(true),
			}

			slaRuleModel := &backuprecoveryv1.SlaRule{
				BackupRunType: core.StringPtr("kIncremental"),
				SlaMinutes:    core.Int64Ptr(int64(1)),
			}

			keyValuePairModel := &backuprecoveryv1.KeyValuePair{
				Key:   core.StringPtr("configKey"),
				Value: core.StringPtr("configValue"),
			}

			physicalVolumeProtectionGroupObjectParamsModel := &backuprecoveryv1.PhysicalVolumeProtectionGroupObjectParams{
				ID:                 core.Int64Ptr(int64(3)),
				VolumeGuids:        []string{"volumeGuid1"},
				EnableSystemBackup: core.BoolPtr(true),
				ExcludedVssWriters: []string{"writerName1", "writerName2"},
			}

			indexingPolicyModel := &backuprecoveryv1.IndexingPolicy{
				EnableIndexing: core.BoolPtr(true),
				IncludePaths:   []string{"~/dir1"},
				ExcludePaths:   []string{"~/dir2"},
			}

			commonPreBackupScriptParamsModel := &backuprecoveryv1.CommonPreBackupScriptParams{
				Path:            core.StringPtr("~/script1"),
				Params:          core.StringPtr("param1"),
				TimeoutSecs:     core.Int64Ptr(int64(1)),
				IsActive:        core.BoolPtr(true),
				ContinueOnError: core.BoolPtr(true),
			}

			commonPostBackupScriptParamsModel := &backuprecoveryv1.CommonPostBackupScriptParams{
				Path:        core.StringPtr("~/script2"),
				Params:      core.StringPtr("param2"),
				TimeoutSecs: core.Int64Ptr(int64(1)),
				IsActive:    core.BoolPtr(true),
			}

			prePostScriptParamsModel := &backuprecoveryv1.PrePostScriptParams{
				PreScript:  commonPreBackupScriptParamsModel,
				PostScript: commonPostBackupScriptParamsModel,
			}

			physicalVolumeProtectionGroupParamsModel := &backuprecoveryv1.PhysicalVolumeProtectionGroupParams{
				Objects:                        []backuprecoveryv1.PhysicalVolumeProtectionGroupObjectParams{*physicalVolumeProtectionGroupObjectParamsModel},
				IndexingPolicy:                 indexingPolicyModel,
				PerformSourceSideDeduplication: core.BoolPtr(true),
				Quiesce:                        core.BoolPtr(true),
				ContinueOnQuiesceFailure:       core.BoolPtr(true),
				IncrementalBackupAfterRestart:  core.BoolPtr(true),
				PrePostScript:                  prePostScriptParamsModel,
				DedupExclusionSourceIds:        []int64{int64(26)},
				ExcludedVssWriters:             []string{"writerName1", "writerName2"},
				CobmrBackup:                    core.BoolPtr(true),
			}

			physicalFileBackupPathParamsModel := &backuprecoveryv1.PhysicalFileBackupPathParams{
				IncludedPath:      core.StringPtr("~/dir1/"),
				ExcludedPaths:     []string{"~/dir2"},
				SkipNestedVolumes: core.BoolPtr(true),
			}

			physicalFileProtectionGroupObjectParamsModel := &backuprecoveryv1.PhysicalFileProtectionGroupObjectParams{
				ExcludedVssWriters:                   []string{"writerName1", "writerName2"},
				ID:                                   core.Int64Ptr(int64(2)),
				FilePaths:                            []backuprecoveryv1.PhysicalFileBackupPathParams{*physicalFileBackupPathParamsModel},
				UsesPathLevelSkipNestedVolumeSetting: core.BoolPtr(true),
				NestedVolumeTypesToSkip:              []string{"volume1"},
				FollowNasSymlinkTarget:               core.BoolPtr(true),
				MetadataFilePath:                     core.StringPtr("~/dir3"),
			}

			cancellationTimeoutParamsModel := &backuprecoveryv1.CancellationTimeoutParams{
				TimeoutMins: core.Int64Ptr(int64(26)),
				BackupType:  core.StringPtr("kRegular"),
			}

			physicalFileProtectionGroupParamsModel := &backuprecoveryv1.PhysicalFileProtectionGroupParams{
				ExcludedVssWriters:             []string{"writerName1", "writerName2"},
				Objects:                        []backuprecoveryv1.PhysicalFileProtectionGroupObjectParams{*physicalFileProtectionGroupObjectParamsModel},
				IndexingPolicy:                 indexingPolicyModel,
				PerformSourceSideDeduplication: core.BoolPtr(true),
				PerformBrickBasedDeduplication: core.BoolPtr(true),
				TaskTimeouts:                   []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				Quiesce:                        core.BoolPtr(true),
				ContinueOnQuiesceFailure:       core.BoolPtr(true),
				CobmrBackup:                    core.BoolPtr(true),
				PrePostScript:                  prePostScriptParamsModel,
				DedupExclusionSourceIds:        []int64{int64(26)},
				GlobalExcludePaths:             []string{"~/dir1"},
				GlobalExcludeFS:                []string{"~/dir2"},
				IgnorableErrors:                []string{"kEOF"},
				AllowParallelRuns:              core.BoolPtr(true),
			}

			physicalProtectionGroupParamsModel := &backuprecoveryv1.PhysicalProtectionGroupParams{
				ProtectionType:             core.StringPtr("kFile"),
				VolumeProtectionTypeParams: physicalVolumeProtectionGroupParamsModel,
				FileProtectionTypeParams:   physicalFileProtectionGroupParamsModel,
			}

			advancedSettingsModel := &backuprecoveryv1.AdvancedSettings{
				ClonedDbBackupStatus:            core.StringPtr("kError"),
				DbBackupIfNotOnlineStatus:       core.StringPtr("kError"),
				MissingDbBackupStatus:           core.StringPtr("kError"),
				OfflineRestoringDbBackupStatus:  core.StringPtr("kError"),
				ReadOnlyDbBackupStatus:          core.StringPtr("kError"),
				ReportAllNonAutoprotectDbErrors: core.StringPtr("kError"),
			}

			filterModel := &backuprecoveryv1.Filter{
				FilterString:        core.StringPtr("filterString"),
				IsRegularExpression: core.BoolPtr(false),
			}

			mssqlFileProtectionGroupHostParamsModel := &backuprecoveryv1.MSSQLFileProtectionGroupHostParams{
				DisableSourceSideDeduplication: core.BoolPtr(true),
				HostID:                         core.Int64Ptr(int64(26)),
			}

			mssqlFileProtectionGroupObjectParamsModel := &backuprecoveryv1.MSSQLFileProtectionGroupObjectParams{
				ID: core.Int64Ptr(int64(6)),
			}

			mssqlFileProtectionGroupParamsModel := &backuprecoveryv1.MSSQLFileProtectionGroupParams{
				AagBackupPreferenceType:        core.StringPtr("kPrimaryReplicaOnly"),
				AdvancedSettings:               advancedSettingsModel,
				BackupSystemDbs:                core.BoolPtr(true),
				ExcludeFilters:                 []backuprecoveryv1.Filter{*filterModel},
				FullBackupsCopyOnly:            core.BoolPtr(true),
				LogBackupNumStreams:            core.Int64Ptr(int64(38)),
				LogBackupWithClause:            core.StringPtr("backupWithClause"),
				PrePostScript:                  prePostScriptParamsModel,
				UseAagPreferencesFromServer:    core.BoolPtr(true),
				UserDbBackupPreferenceType:     core.StringPtr("kBackupAllDatabases"),
				AdditionalHostParams:           []backuprecoveryv1.MSSQLFileProtectionGroupHostParams{*mssqlFileProtectionGroupHostParamsModel},
				Objects:                        []backuprecoveryv1.MSSQLFileProtectionGroupObjectParams{*mssqlFileProtectionGroupObjectParamsModel},
				PerformSourceSideDeduplication: core.BoolPtr(true),
			}

			mssqlNativeProtectionGroupObjectParamsModel := &backuprecoveryv1.MSSQLNativeProtectionGroupObjectParams{
				ID: core.Int64Ptr(int64(6)),
			}

			mssqlNativeProtectionGroupParamsModel := &backuprecoveryv1.MSSQLNativeProtectionGroupParams{
				AagBackupPreferenceType:     core.StringPtr("kPrimaryReplicaOnly"),
				AdvancedSettings:            advancedSettingsModel,
				BackupSystemDbs:             core.BoolPtr(true),
				ExcludeFilters:              []backuprecoveryv1.Filter{*filterModel},
				FullBackupsCopyOnly:         core.BoolPtr(true),
				LogBackupNumStreams:         core.Int64Ptr(int64(38)),
				LogBackupWithClause:         core.StringPtr("backupWithClause"),
				PrePostScript:               prePostScriptParamsModel,
				UseAagPreferencesFromServer: core.BoolPtr(true),
				UserDbBackupPreferenceType:  core.StringPtr("kBackupAllDatabases"),
				NumStreams:                  core.Int64Ptr(int64(38)),
				Objects:                     []backuprecoveryv1.MSSQLNativeProtectionGroupObjectParams{*mssqlNativeProtectionGroupObjectParamsModel},
				WithClause:                  core.StringPtr("withClause"),
			}

			mssqlVolumeProtectionGroupHostParamsModel := &backuprecoveryv1.MSSQLVolumeProtectionGroupHostParams{
				EnableSystemBackup: core.BoolPtr(true),
				HostID:             core.Int64Ptr(int64(8)),
				VolumeGuids:        []string{"volumeGuid1"},
			}

			mssqlVolumeProtectionGroupObjectParamsModel := &backuprecoveryv1.MSSQLVolumeProtectionGroupObjectParams{
				ID: core.Int64Ptr(int64(6)),
			}

			mssqlVolumeProtectionGroupParamsModel := &backuprecoveryv1.MSSQLVolumeProtectionGroupParams{
				AagBackupPreferenceType:       core.StringPtr("kPrimaryReplicaOnly"),
				AdvancedSettings:              advancedSettingsModel,
				BackupSystemDbs:               core.BoolPtr(true),
				ExcludeFilters:                []backuprecoveryv1.Filter{*filterModel},
				FullBackupsCopyOnly:           core.BoolPtr(true),
				LogBackupNumStreams:           core.Int64Ptr(int64(38)),
				LogBackupWithClause:           core.StringPtr("backupWithClause"),
				PrePostScript:                 prePostScriptParamsModel,
				UseAagPreferencesFromServer:   core.BoolPtr(true),
				UserDbBackupPreferenceType:    core.StringPtr("kBackupAllDatabases"),
				AdditionalHostParams:          []backuprecoveryv1.MSSQLVolumeProtectionGroupHostParams{*mssqlVolumeProtectionGroupHostParamsModel},
				BackupDbVolumesOnly:           core.BoolPtr(true),
				IncrementalBackupAfterRestart: core.BoolPtr(true),
				IndexingPolicy:                indexingPolicyModel,
				Objects:                       []backuprecoveryv1.MSSQLVolumeProtectionGroupObjectParams{*mssqlVolumeProtectionGroupObjectParamsModel},
			}

			mssqlProtectionGroupParamsModel := &backuprecoveryv1.MSSQLProtectionGroupParams{
				FileProtectionTypeParams:   mssqlFileProtectionGroupParamsModel,
				NativeProtectionTypeParams: mssqlNativeProtectionGroupParamsModel,
				ProtectionType:             core.StringPtr("kFile"),
				VolumeProtectionTypeParams: mssqlVolumeProtectionGroupParamsModel,
			}

			createProtectionGroupOptions := &backuprecoveryv1.CreateProtectionGroupOptions{
				XIBMTenantID:               core.StringPtr("tenantId"),
				Name:                       core.StringPtr("create-protection-group"),
				PolicyID:                   core.StringPtr("xxxxxxxxxxxxxxxx:xxxxxxxxxxxxx:xx"),
				Environment:                core.StringPtr("kPhysical"),
				Priority:                   core.StringPtr("kLow"),
				Description:                core.StringPtr("Protection Group"),
				StartTime:                  timeOfDayModel,
				EndTimeUsecs:               core.Int64Ptr(int64(26)),
				LastModifiedTimestampUsecs: core.Int64Ptr(int64(26)),
				AlertPolicy:                protectionGroupAlertingPolicyModel,
				Sla:                        []backuprecoveryv1.SlaRule{*slaRuleModel},
				QosPolicy:                  core.StringPtr("kBackupHDD"),
				AbortInBlackouts:           core.BoolPtr(true),
				PauseInBlackouts:           core.BoolPtr(true),
				IsPaused:                   core.BoolPtr(true),
				AdvancedConfigs:            []backuprecoveryv1.KeyValuePair{*keyValuePairModel},
				PhysicalParams:             physicalProtectionGroupParamsModel,
				MssqlParams:                mssqlProtectionGroupParamsModel,
			}

			protectionGroupResponse, response, err := backupRecoveryService.CreateProtectionGroup(createProtectionGroupOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(201))
			Expect(protectionGroupResponse).ToNot(BeNil())

			protectionGroupIdLink = *protectionGroupResponse.ID
			fmt.Fprintf(GinkgoWriter, "Saved protectionGroupIdLink value: %v\n", protectionGroupIdLink)
		})
	})

	Describe(`CreateRecovery - Performs a Recovery`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`CreateRecovery(createRecoveryOptions *CreateRecoveryOptions)`, func() {
			commonRecoverObjectSnapshotParamsModel := &backuprecoveryv1.CommonRecoverObjectSnapshotParams{
				SnapshotID:          core.StringPtr("snapshotID"),
				PointInTimeUsecs:    core.Int64Ptr(int64(26)),
				ProtectionGroupID:   core.StringPtr("protectionGroupID"),
				ProtectionGroupName: core.StringPtr("protectionGroupName"),
				RecoverFromStandby:  core.BoolPtr(true),
			}

			physicalTargetParamsForRecoverVolumeMountTargetModel := &backuprecoveryv1.PhysicalTargetParamsForRecoverVolumeMountTarget{
				ID: core.Int64Ptr(int64(26)),
			}

			recoverVolumeMappingModel := &backuprecoveryv1.RecoverVolumeMapping{
				SourceVolumeGuid:      core.StringPtr("sourceVolumeGuid"),
				DestinationVolumeGuid: core.StringPtr("destinationVolumeGuid"),
			}

			physicalTargetParamsForRecoverVolumeVlanConfigModel := &backuprecoveryv1.PhysicalTargetParamsForRecoverVolumeVlanConfig{
				ID:          core.Int64Ptr(int64(38)),
				DisableVlan: core.BoolPtr(true),
			}

			recoverPhysicalVolumeParamsPhysicalTargetParamsModel := &backuprecoveryv1.RecoverPhysicalVolumeParamsPhysicalTargetParams{
				MountTarget:        physicalTargetParamsForRecoverVolumeMountTargetModel,
				VolumeMapping:      []backuprecoveryv1.RecoverVolumeMapping{*recoverVolumeMappingModel},
				ForceUnmountVolume: core.BoolPtr(true),
				VlanConfig:         physicalTargetParamsForRecoverVolumeVlanConfigModel,
			}

			recoverPhysicalParamsRecoverVolumeParamsModel := &backuprecoveryv1.RecoverPhysicalParamsRecoverVolumeParams{
				TargetEnvironment:    core.StringPtr("kPhysical"),
				PhysicalTargetParams: recoverPhysicalVolumeParamsPhysicalTargetParamsModel,
			}

			physicalMountVolumesOriginalTargetConfigServerCredentialsModel := &backuprecoveryv1.PhysicalMountVolumesOriginalTargetConfigServerCredentials{
				Username: core.StringPtr("Username"),
				Password: core.StringPtr("Password"),
			}

			physicalTargetParamsForMountVolumeOriginalTargetConfigModel := &backuprecoveryv1.PhysicalTargetParamsForMountVolumeOriginalTargetConfig{
				ServerCredentials: physicalMountVolumesOriginalTargetConfigServerCredentialsModel,
			}

			recoverTargetModel := &backuprecoveryv1.RecoverTarget{
				ID: core.Int64Ptr(int64(26)),
			}

			physicalMountVolumesNewTargetConfigServerCredentialsModel := &backuprecoveryv1.PhysicalMountVolumesNewTargetConfigServerCredentials{
				Username: core.StringPtr("Username"),
				Password: core.StringPtr("Password"),
			}

			physicalTargetParamsForMountVolumeNewTargetConfigModel := &backuprecoveryv1.PhysicalTargetParamsForMountVolumeNewTargetConfig{
				MountTarget:       recoverTargetModel,
				ServerCredentials: physicalMountVolumesNewTargetConfigServerCredentialsModel,
			}

			physicalTargetParamsForMountVolumeVlanConfigModel := &backuprecoveryv1.PhysicalTargetParamsForMountVolumeVlanConfig{
				ID:          core.Int64Ptr(int64(38)),
				DisableVlan: core.BoolPtr(true),
			}

			mountPhysicalVolumeParamsPhysicalTargetParamsModel := &backuprecoveryv1.MountPhysicalVolumeParamsPhysicalTargetParams{
				MountToOriginalTarget: core.BoolPtr(true),
				OriginalTargetConfig:  physicalTargetParamsForMountVolumeOriginalTargetConfigModel,
				NewTargetConfig:       physicalTargetParamsForMountVolumeNewTargetConfigModel,
				ReadOnlyMount:         core.BoolPtr(true),
				VolumeNames:           []string{"volume1"},
				VlanConfig:            physicalTargetParamsForMountVolumeVlanConfigModel,
			}

			recoverPhysicalParamsMountVolumeParamsModel := &backuprecoveryv1.RecoverPhysicalParamsMountVolumeParams{
				TargetEnvironment:    core.StringPtr("kPhysical"),
				PhysicalTargetParams: mountPhysicalVolumeParamsPhysicalTargetParamsModel,
			}

			commonRecoverFileAndFolderInfoModel := &backuprecoveryv1.CommonRecoverFileAndFolderInfo{
				AbsolutePath:       core.StringPtr("~/folder1"),
				IsDirectory:        core.BoolPtr(true),
				IsViewFileRecovery: core.BoolPtr(true),
			}

			physicalTargetParamsForRecoverFileAndFolderRecoverTargetModel := &backuprecoveryv1.PhysicalTargetParamsForRecoverFileAndFolderRecoverTarget{
				ID: core.Int64Ptr(int64(26)),
			}

			physicalTargetParamsForRecoverFileAndFolderVlanConfigModel := &backuprecoveryv1.PhysicalTargetParamsForRecoverFileAndFolderVlanConfig{
				ID:          core.Int64Ptr(int64(38)),
				DisableVlan: core.BoolPtr(true),
			}

			recoverPhysicalFileAndFolderParamsPhysicalTargetParamsModel := &backuprecoveryv1.RecoverPhysicalFileAndFolderParamsPhysicalTargetParams{
				RecoverTarget:             physicalTargetParamsForRecoverFileAndFolderRecoverTargetModel,
				RestoreToOriginalPaths:    core.BoolPtr(true),
				OverwriteExisting:         core.BoolPtr(true),
				AlternateRestoreDirectory: core.StringPtr("~/dirAlt"),
				PreserveAttributes:        core.BoolPtr(true),
				PreserveTimestamps:        core.BoolPtr(true),
				PreserveAcls:              core.BoolPtr(true),
				ContinueOnError:           core.BoolPtr(true),
				SaveSuccessFiles:          core.BoolPtr(true),
				VlanConfig:                physicalTargetParamsForRecoverFileAndFolderVlanConfigModel,
				RestoreEntityType:         core.StringPtr("kRegular"),
			}

			recoverPhysicalParamsRecoverFileAndFolderParamsModel := &backuprecoveryv1.RecoverPhysicalParamsRecoverFileAndFolderParams{
				FilesAndFolders:      []backuprecoveryv1.CommonRecoverFileAndFolderInfo{*commonRecoverFileAndFolderInfoModel},
				TargetEnvironment:    core.StringPtr("kPhysical"),
				PhysicalTargetParams: recoverPhysicalFileAndFolderParamsPhysicalTargetParamsModel,
			}

			recoverPhysicalParamsDownloadFileAndFolderParamsModel := &backuprecoveryv1.RecoverPhysicalParamsDownloadFileAndFolderParams{
				ExpiryTimeUsecs:  core.Int64Ptr(int64(26)),
				FilesAndFolders:  []backuprecoveryv1.CommonRecoverFileAndFolderInfo{*commonRecoverFileAndFolderInfoModel},
				DownloadFilePath: core.StringPtr("~/downloadFile"),
			}

			recoverPhysicalParamsSystemRecoveryParamsModel := &backuprecoveryv1.RecoverPhysicalParamsSystemRecoveryParams{
				FullNasPath: core.StringPtr("~/nas"),
			}

			recoverPhysicalParamsModel := &backuprecoveryv1.RecoverPhysicalParams{
				Objects:                     []backuprecoveryv1.CommonRecoverObjectSnapshotParams{*commonRecoverObjectSnapshotParamsModel},
				RecoveryAction:              core.StringPtr("RecoverPhysicalVolumes"),
				RecoverVolumeParams:         recoverPhysicalParamsRecoverVolumeParamsModel,
				MountVolumeParams:           recoverPhysicalParamsMountVolumeParamsModel,
				RecoverFileAndFolderParams:  recoverPhysicalParamsRecoverFileAndFolderParamsModel,
				DownloadFileAndFolderParams: recoverPhysicalParamsDownloadFileAndFolderParamsModel,
				SystemRecoveryParams:        recoverPhysicalParamsSystemRecoveryParamsModel,
			}

			aagInfoModel := &backuprecoveryv1.AAGInfo{
				Name:     core.StringPtr("aagInfoName"),
				ObjectID: core.Int64Ptr(int64(26)),
			}

			hostInformationModel := &backuprecoveryv1.HostInformation{
				ID:          core.StringPtr("hostInfoId"),
				Name:        core.StringPtr("hostInfoName"),
				Environment: core.StringPtr("kPhysical"),
			}

			multiStageRestoreOptionsModel := &backuprecoveryv1.MultiStageRestoreOptions{
				EnableAutoSync:          core.BoolPtr(true),
				EnableMultiStageRestore: core.BoolPtr(true),
			}

			filenamePatternToDirectoryModel := &backuprecoveryv1.FilenamePatternToDirectory{
				Directory:       core.StringPtr("~/dir1"),
				FilenamePattern: core.StringPtr(".sql"),
			}

			recoveryObjectIdentifierModel := &backuprecoveryv1.RecoveryObjectIdentifier{
				ID: core.Int64Ptr(int64(26)),
			}

			recoverSqlAppNewSourceConfigModel := &backuprecoveryv1.RecoverSqlAppNewSourceConfig{
				KeepCdc:                     core.BoolPtr(true),
				MultiStageRestoreOptions:    multiStageRestoreOptionsModel,
				NativeLogRecoveryWithClause: core.StringPtr("LogRecoveryWithClause"),
				NativeRecoveryWithClause:    core.StringPtr("RecoveryWithClause"),
				OverwritingPolicy:           core.StringPtr("FailIfExists"),
				ReplayEntireLastLog:         core.BoolPtr(true),
				RestoreTimeUsecs:            core.Int64Ptr(int64(26)),
				SecondaryDataFilesDirList:   []backuprecoveryv1.FilenamePatternToDirectory{*filenamePatternToDirectoryModel},
				WithNoRecovery:              core.BoolPtr(true),
				DataFileDirectoryLocation:   core.StringPtr("~/dir1"),
				DatabaseName:                core.StringPtr("recovery-database-sql"),
				Host:                        recoveryObjectIdentifierModel,
				InstanceName:                core.StringPtr("database-instance-1"),
				LogFileDirectoryLocation:    core.StringPtr("~/dir2"),
			}

			recoverSqlAppOriginalSourceConfigModel := &backuprecoveryv1.RecoverSqlAppOriginalSourceConfig{
				KeepCdc:                     core.BoolPtr(true),
				MultiStageRestoreOptions:    multiStageRestoreOptionsModel,
				NativeLogRecoveryWithClause: core.StringPtr("LogRecoveryWithClause"),
				NativeRecoveryWithClause:    core.StringPtr("RecoveryWithClause"),
				OverwritingPolicy:           core.StringPtr("FailIfExists"),
				ReplayEntireLastLog:         core.BoolPtr(true),
				RestoreTimeUsecs:            core.Int64Ptr(int64(26)),
				SecondaryDataFilesDirList:   []backuprecoveryv1.FilenamePatternToDirectory{*filenamePatternToDirectoryModel},
				WithNoRecovery:              core.BoolPtr(true),
				CaptureTailLogs:             core.BoolPtr(true),
				DataFileDirectoryLocation:   core.StringPtr("~/dir1"),
				LogFileDirectoryLocation:    core.StringPtr("~/dir2"),
				NewDatabaseName:             core.StringPtr("recovery-database-sql-new"),
			}

			sqlTargetParamsForRecoverSqlAppModel := &backuprecoveryv1.SqlTargetParamsForRecoverSqlApp{
				NewSourceConfig:      recoverSqlAppNewSourceConfigModel,
				OriginalSourceConfig: recoverSqlAppOriginalSourceConfigModel,
				RecoverToNewSource:   core.BoolPtr(true),
			}

			recoverSqlAppParamsModel := &backuprecoveryv1.RecoverSqlAppParams{
				SnapshotID:          core.StringPtr("snapshotId"),
				PointInTimeUsecs:    core.Int64Ptr(int64(26)),
				ProtectionGroupID:   core.StringPtr("protectionGroupId"),
				ProtectionGroupName: core.StringPtr("protectionGroupName"),
				RecoverFromStandby:  core.BoolPtr(true),
				AagInfo:             aagInfoModel,
				HostInfo:            hostInformationModel,
				IsEncrypted:         core.BoolPtr(true),
				SqlTargetParams:     sqlTargetParamsForRecoverSqlAppModel,
				TargetEnvironment:   core.StringPtr("kSQL"),
			}

			recoveryVlanConfigModel := &backuprecoveryv1.RecoveryVlanConfig{
				ID:          core.Int64Ptr(int64(38)),
				DisableVlan: core.BoolPtr(true),
			}

			recoverSqlParamsModel := &backuprecoveryv1.RecoverSqlParams{
				RecoverAppParams: []backuprecoveryv1.RecoverSqlAppParams{*recoverSqlAppParamsModel},
				RecoveryAction:   core.StringPtr("RecoverApps"),
				VlanConfig:       recoveryVlanConfigModel,
			}

			createRecoveryOptions := &backuprecoveryv1.CreateRecoveryOptions{
				XIBMTenantID:         core.StringPtr("tenantId"),
				Name:                 core.StringPtr("create-recovery"),
				SnapshotEnvironment:  core.StringPtr("kPhysical"),
				PhysicalParams:       recoverPhysicalParamsModel,
				MssqlParams:          recoverSqlParamsModel,
				RequestInitiatorType: core.StringPtr("UIUser"),
			}

			recovery, response, err := backupRecoveryService.CreateRecovery(createRecoveryOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(201))
			Expect(recovery).ToNot(BeNil())

			recoveryIdLink = *recovery.ID
			fmt.Fprintf(GinkgoWriter, "Saved recoveryIdLink value: %v\n", recoveryIdLink)
		})
	})

	Describe(`RegisterProtectionSource - Register a Protection Source`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`RegisterProtectionSource(registerProtectionSourceOptions *RegisterProtectionSourceOptions)`, func() {
			connectionConfigModel := &backuprecoveryv1.ConnectionConfig{
				ConnectionID:           core.Int64Ptr(int64(26)),
				EntityID:               core.Int64Ptr(int64(26)),
				ConnectorGroupID:       core.Int64Ptr(int64(26)),
				DataSourceConnectionID: core.StringPtr("DatasourceConnectionId"),
			}

			keyValuePairModel := &backuprecoveryv1.KeyValuePair{
				Key:   core.StringPtr("configKey"),
				Value: core.StringPtr("configValue"),
			}

			physicalSourceRegistrationParamsModel := &backuprecoveryv1.PhysicalSourceRegistrationParams{
				Endpoint:      core.StringPtr("xxx.xx.xx.xx"),
				ForceRegister: core.BoolPtr(true),
				HostType:      core.StringPtr("kLinux"),
				PhysicalType:  core.StringPtr("kGroup"),
				Applications:  []string{"kSQL"},
			}

			registerProtectionSourceOptions := &backuprecoveryv1.RegisterProtectionSourceOptions{
				XIBMTenantID:           core.StringPtr("tenantId"),
				Environment:            core.StringPtr("kPhysical"),
				Name:                   core.StringPtr("register-protection-source"),
				IsInternalEncrypted:    core.BoolPtr(true),
				EncryptionKey:          core.StringPtr("encryptionKey"),
				ConnectionID:           core.Int64Ptr(int64(26)),
				Connections:            []backuprecoveryv1.ConnectionConfig{*connectionConfigModel},
				ConnectorGroupID:       core.Int64Ptr(int64(26)),
				AdvancedConfigs:        []backuprecoveryv1.KeyValuePair{*keyValuePairModel},
				DataSourceConnectionID: core.StringPtr("DatasourceConnectionId"),
				PhysicalParams:         physicalSourceRegistrationParamsModel,
			}

			sourceRegistrationResponseParams, response, err := backupRecoveryService.RegisterProtectionSource(registerProtectionSourceOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(201))
			Expect(sourceRegistrationResponseParams).ToNot(BeNil())

			protectionSourceIdLink = *sourceRegistrationResponseParams.ID
			fmt.Fprintf(GinkgoWriter, "Saved protectionSourceIdLink value: %v\n", protectionSourceIdLink)
		})
	})

	Describe(`CreateAccessToken - Create Access Token`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`CreateAccessToken(createAccessTokenOptions *CreateAccessTokenOptions)`, func() {
			createAccessTokenOptions := &backuprecoveryv1.CreateAccessTokenOptions{
				Username: core.StringPtr("testString"),
				Password: core.StringPtr("testString"),
				Domain:   core.StringPtr("testString"),
			}

			tokenResponse, response, err := backupRecoveryService.CreateAccessToken(createAccessTokenOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(201))
			Expect(tokenResponse).ToNot(BeNil())
		})
	})

	Describe(`DownloadAgent - Download agent`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`DownloadAgent(downloadAgentOptions *DownloadAgentOptions)`, func() {
			linuxAgentParamsModel := &backuprecoveryv1.LinuxAgentParams{
				PackageType: core.StringPtr("kScript"),
			}

			downloadAgentOptions := &backuprecoveryv1.DownloadAgentOptions{
				XIBMTenantID: core.StringPtr("tenantId"),
				Platform:     core.StringPtr("kWindows"),
				LinuxParams:  linuxAgentParamsModel,
			}

			result, response, err := backupRecoveryService.DownloadAgent(downloadAgentOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(result).ToNot(BeNil())
		})
	})

	Describe(`GetUpgradeTasks - Get upgrade tasks`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetUpgradeTasks(getUpgradeTasksOptions *GetUpgradeTasksOptions)`, func() {
			getUpgradeTasksOptions := &backuprecoveryv1.GetUpgradeTasksOptions{
				XIBMTenantID: core.StringPtr("tenantId"),
				Ids:          []int64{int64(26)},
			}

			agentUpgradeTaskStates, response, err := backupRecoveryService.GetUpgradeTasks(getUpgradeTasksOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(agentUpgradeTaskStates).ToNot(BeNil())
		})
	})

	Describe(`CreateUpgradeTask - Create an upgrade task`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`CreateUpgradeTask(createUpgradeTaskOptions *CreateUpgradeTaskOptions)`, func() {
			createUpgradeTaskOptions := &backuprecoveryv1.CreateUpgradeTaskOptions{
				XIBMTenantID:         core.StringPtr("tenantId"),
				AgentIDs:             []int64{int64(26)},
				Description:          core.StringPtr("Upgrade task"),
				Name:                 core.StringPtr("create-upgrade-task"),
				RetryTaskID:          core.Int64Ptr(int64(26)),
				ScheduleEndTimeUsecs: core.Int64Ptr(int64(26)),
				ScheduleTimeUsecs:    core.Int64Ptr(int64(26)),
			}

			agentUpgradeTaskState, response, err := backupRecoveryService.CreateUpgradeTask(createUpgradeTaskOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(201))
			Expect(agentUpgradeTaskState).ToNot(BeNil())
		})
	})

	Describe(`ListProtectionSources - List Protection Sources`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`ListProtectionSources(listProtectionSourcesOptions *ListProtectionSourcesOptions)`, func() {
			listProtectionSourcesOptions := &backuprecoveryv1.ListProtectionSourcesOptions{
				XIBMTenantID:                core.StringPtr("tenantId"),
				ExcludeOffice365Types:       []string{"kDomain"},
				GetTeamsChannels:            core.BoolPtr(true),
				AfterCursorEntityID:         core.Int64Ptr(int64(26)),
				BeforeCursorEntityID:        core.Int64Ptr(int64(26)),
				NodeID:                      core.Int64Ptr(int64(26)),
				PageSize:                    core.Int64Ptr(int64(26)),
				HasValidMailbox:             core.BoolPtr(true),
				HasValidOnedrive:            core.BoolPtr(true),
				IsSecurityGroup:             core.BoolPtr(true),
				ID:                          core.Int64Ptr(int64(26)),
				NumLevels:                   core.Float64Ptr(float64(72.5)),
				ExcludeTypes:                []string{"kVCenter"},
				ExcludeAwsTypes:             []string{"kEC2Instance"},
				ExcludeKubernetesTypes:      []string{"kService"},
				IncludeDatastores:           core.BoolPtr(true),
				IncludeNetworks:             core.BoolPtr(true),
				IncludeVMFolders:            core.BoolPtr(true),
				IncludeSfdcFields:           core.BoolPtr(true),
				IncludeSystemVApps:          core.BoolPtr(true),
				Environments:                []string{"kVMware"},
				Environment:                 core.StringPtr("kPhysical"),
				IncludeEntityPermissionInfo: core.BoolPtr(true),
				Sids:                        []string{"sid1"},
				IncludeSourceCredentials:    core.BoolPtr(true),
				EncryptionKey:               core.StringPtr("encryptionKey"),
				IncludeObjectProtectionInfo: core.BoolPtr(true),
				PruneNonCriticalInfo:        core.BoolPtr(true),
				PruneAggregationInfo:        core.BoolPtr(true),
				RequestInitiatorType:        core.StringPtr("requestInitiatorType"),
				UseCachedData:               core.BoolPtr(true),
				AllUnderHierarchy:           core.BoolPtr(true),
			}

			protectionSourceNodes, response, err := backupRecoveryService.ListProtectionSources(listProtectionSourcesOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(protectionSourceNodes).ToNot(BeNil())
		})
	})

	Describe(`GetDataSourceConnections - Get data-source connections`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetDataSourceConnections(getDataSourceConnectionsOptions *GetDataSourceConnectionsOptions)`, func() {
			getDataSourceConnectionsOptions := &backuprecoveryv1.GetDataSourceConnectionsOptions{
				XIBMTenantID:    core.StringPtr("tenantId"),
				ConnectionIds:   []string{"connectionId1", "connectionId2"},
				ConnectionNames: []string{"connectionName1", "connectionName2"},
			}

			dataSourceConnectionList, response, err := backupRecoveryService.GetDataSourceConnections(getDataSourceConnectionsOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(dataSourceConnectionList).ToNot(BeNil())
		})
	})

	Describe(`PatchDataSourceConnection - Patch a data-source connection using its ID`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`PatchDataSourceConnection(patchDataSourceConnectionOptions *PatchDataSourceConnectionOptions)`, func() {
			patchDataSourceConnectionOptions := &backuprecoveryv1.PatchDataSourceConnectionOptions{
				ConnectionID:   &connectionIdLink,
				XIBMTenantID:   core.StringPtr("tenantId"),
				ConnectionName: core.StringPtr("connectionName"),
			}

			dataSourceConnection, response, err := backupRecoveryService.PatchDataSourceConnection(patchDataSourceConnectionOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(dataSourceConnection).ToNot(BeNil())
		})
	})

	Describe(`GenerateDataSourceConnectionRegistrationToken - Generate registration token for a data-source connection`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GenerateDataSourceConnectionRegistrationToken(generateDataSourceConnectionRegistrationTokenOptions *GenerateDataSourceConnectionRegistrationTokenOptions)`, func() {
			generateDataSourceConnectionRegistrationTokenOptions := &backuprecoveryv1.GenerateDataSourceConnectionRegistrationTokenOptions{
				ConnectionID: &connectionIdLink,
				XIBMTenantID: core.StringPtr("tenantId"),
			}

			result, response, err := backupRecoveryService.GenerateDataSourceConnectionRegistrationToken(generateDataSourceConnectionRegistrationTokenOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(result).ToNot(BeNil())
		})
	})

	Describe(`GetDataSourceConnectors - Get data-source connectors`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetDataSourceConnectors(getDataSourceConnectorsOptions *GetDataSourceConnectorsOptions)`, func() {
			getDataSourceConnectorsOptions := &backuprecoveryv1.GetDataSourceConnectorsOptions{
				XIBMTenantID:   core.StringPtr("tenantId"),
				ConnectorIds:   []string{"connectorId1", "connectorId2"},
				ConnectorNames: []string{"connectionName1", "connectionName2"},
				ConnectionID:   &connectionIdLink,
			}

			dataSourceConnectorList, response, err := backupRecoveryService.GetDataSourceConnectors(getDataSourceConnectorsOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(dataSourceConnectorList).ToNot(BeNil())
		})
	})

	Describe(`GetConnectorMetadata - Get information about the available connectors`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetConnectorMetadata(getConnectorMetadataOptions *GetConnectorMetadataOptions)`, func() {
			getConnectorMetadataOptions := &backuprecoveryv1.GetConnectorMetadataOptions{
				XIBMTenantID: core.StringPtr("tenantId"),
			}

			connectorMetadata, response, err := backupRecoveryService.GetConnectorMetadata(getConnectorMetadataOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(connectorMetadata).ToNot(BeNil())
		})
	})

	Describe(`PatchDataSourceConnector - Patch a data-source connector using its ID`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`PatchDataSourceConnector(patchDataSourceConnectorOptions *PatchDataSourceConnectorOptions)`, func() {
			patchDataSourceConnectorOptions := &backuprecoveryv1.PatchDataSourceConnectorOptions{
				ConnectorID:   core.StringPtr("connectorID"),
				XIBMTenantID:  core.StringPtr("tenantId"),
				ConnectorName: core.StringPtr("connectorName"),
			}

			dataSourceConnector, response, err := backupRecoveryService.PatchDataSourceConnector(patchDataSourceConnectorOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(dataSourceConnector).ToNot(BeNil())
		})
	})

	Describe(`GetDataSourceConnectorLogs - Lists the data-source connector logs`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetDataSourceConnectorLogs(getDataSourceConnectorLogsOptions *GetDataSourceConnectorLogsOptions)`, func() {
			getDataSourceConnectorLogsOptions := &backuprecoveryv1.GetDataSourceConnectorLogsOptions{}

			dataSourceConnectorLogs, response, err := backupRecoveryService.GetDataSourceConnectorLogs(getDataSourceConnectorLogsOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(dataSourceConnectorLogs).ToNot(BeNil())
		})
	})

	Describe(`RegisterDataSourceConnector - Register a data-source connector`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`RegisterDataSourceConnector(registerDataSourceConnectorOptions *RegisterDataSourceConnectorOptions)`, func() {
			registerDataSourceConnectorOptions := &backuprecoveryv1.RegisterDataSourceConnectorOptions{
				RegistrationToken: core.StringPtr("testString"),
				ConnectorID:       core.Int64Ptr(int64(26)),
			}

			response, err := backupRecoveryService.RegisterDataSourceConnector(registerDataSourceConnectorOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(204))
		})
	})

	Describe(`GetDataSourceConnectorStatus - Lists the data-source connector status`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetDataSourceConnectorStatus(getDataSourceConnectorStatusOptions *GetDataSourceConnectorStatusOptions)`, func() {
			getDataSourceConnectorStatusOptions := &backuprecoveryv1.GetDataSourceConnectorStatusOptions{}

			dataSourceConnectorLocalStatus, response, err := backupRecoveryService.GetDataSourceConnectorStatus(getDataSourceConnectorStatusOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(dataSourceConnectorLocalStatus).ToNot(BeNil())
		})
	})

	Describe(`GetObjectSnapshots - List the snapshots for a given object`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetObjectSnapshots(getObjectSnapshotsOptions *GetObjectSnapshotsOptions)`, func() {
			getObjectSnapshotsOptions := &backuprecoveryv1.GetObjectSnapshotsOptions{
				ID:                    core.Int64Ptr(int64(26)),
				XIBMTenantID:          core.StringPtr("tenantId"),
				FromTimeUsecs:         core.Int64Ptr(int64(26)),
				ToTimeUsecs:           core.Int64Ptr(int64(26)),
				RunStartFromTimeUsecs: core.Int64Ptr(int64(26)),
				RunStartToTimeUsecs:   core.Int64Ptr(int64(26)),
				SnapshotActions:       []string{"RecoverPhysicalVolumes"},
				RunTypes:              []string{"kRegular"},
				ProtectionGroupIds:    []string{"protectionGroupId1"},
				RunInstanceIds:        []int64{int64(26)},
				RegionIds:             []string{"regionId1"},
				ObjectActionKeys:      []string{"kVMware"},
			}

			getObjectSnapshotsResponse, response, err := backupRecoveryService.GetObjectSnapshots(getObjectSnapshotsOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(getObjectSnapshotsResponse).ToNot(BeNil())
		})
	})

	Describe(`GetProtectionPolicies - List Protection Policies based on provided filtering parameters`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetProtectionPolicies(getProtectionPoliciesOptions *GetProtectionPoliciesOptions)`, func() {
			getProtectionPoliciesOptions := &backuprecoveryv1.GetProtectionPoliciesOptions{
				XIBMTenantID:              core.StringPtr("tenantId"),
				RequestInitiatorType:      core.StringPtr("UIUser"),
				Ids:                       []string{"policyId1"},
				PolicyNames:               []string{"policyName1"},
				Types:                     []string{"Regular"},
				ExcludeLinkedPolicies:     core.BoolPtr(true),
				IncludeReplicatedPolicies: core.BoolPtr(true),
				IncludeStats:              core.BoolPtr(true),
			}

			protectionPoliciesResponse, response, err := backupRecoveryService.GetProtectionPolicies(getProtectionPoliciesOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(protectionPoliciesResponse).ToNot(BeNil())
		})
	})

	Describe(`GetProtectionPolicyByID - List details about a single Protection Policy`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetProtectionPolicyByID(getProtectionPolicyByIdOptions *GetProtectionPolicyByIdOptions)`, func() {
			getProtectionPolicyByIdOptions := &backuprecoveryv1.GetProtectionPolicyByIdOptions{
				ID:                   &policyIdLink,
				XIBMTenantID:         core.StringPtr("tenantId"),
				RequestInitiatorType: core.StringPtr("UIUser"),
			}

			protectionPolicyResponse, response, err := backupRecoveryService.GetProtectionPolicyByID(getProtectionPolicyByIdOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(protectionPolicyResponse).ToNot(BeNil())
		})
	})

	Describe(`UpdateProtectionPolicy - Update a Protection Policy`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`UpdateProtectionPolicy(updateProtectionPolicyOptions *UpdateProtectionPolicyOptions)`, func() {
			minuteScheduleModel := &backuprecoveryv1.MinuteSchedule{
				Frequency: core.Int64Ptr(int64(1)),
			}

			hourScheduleModel := &backuprecoveryv1.HourSchedule{
				Frequency: core.Int64Ptr(int64(1)),
			}

			dayScheduleModel := &backuprecoveryv1.DaySchedule{
				Frequency: core.Int64Ptr(int64(1)),
			}

			weekScheduleModel := &backuprecoveryv1.WeekSchedule{
				DayOfWeek: []string{"Sunday"},
			}

			monthScheduleModel := &backuprecoveryv1.MonthSchedule{
				DayOfWeek:   []string{"Sunday"},
				WeekOfMonth: core.StringPtr("First"),
				DayOfMonth:  core.Int64Ptr(int64(10)),
			}

			yearScheduleModel := &backuprecoveryv1.YearSchedule{
				DayOfYear: core.StringPtr("First"),
			}

			incrementalScheduleModel := &backuprecoveryv1.IncrementalSchedule{
				Unit:           core.StringPtr("Minutes"),
				MinuteSchedule: minuteScheduleModel,
				HourSchedule:   hourScheduleModel,
				DaySchedule:    dayScheduleModel,
				WeekSchedule:   weekScheduleModel,
				MonthSchedule:  monthScheduleModel,
				YearSchedule:   yearScheduleModel,
			}

			incrementalBackupPolicyModel := &backuprecoveryv1.IncrementalBackupPolicy{
				Schedule: incrementalScheduleModel,
			}

			fullScheduleModel := &backuprecoveryv1.FullSchedule{
				Unit:          core.StringPtr("Days"),
				DaySchedule:   dayScheduleModel,
				WeekSchedule:  weekScheduleModel,
				MonthSchedule: monthScheduleModel,
				YearSchedule:  yearScheduleModel,
			}

			fullBackupPolicyModel := &backuprecoveryv1.FullBackupPolicy{
				Schedule: fullScheduleModel,
			}

			dataLockConfigModel := &backuprecoveryv1.DataLockConfig{
				Mode:                       core.StringPtr("Compliance"),
				Unit:                       core.StringPtr("Days"),
				Duration:                   core.Int64Ptr(int64(1)),
				EnableWormOnExternalTarget: core.BoolPtr(true),
			}

			retentionModel := &backuprecoveryv1.Retention{
				Unit:           core.StringPtr("Days"),
				Duration:       core.Int64Ptr(int64(1)),
				DataLockConfig: dataLockConfigModel,
			}

			fullScheduleAndRetentionModel := &backuprecoveryv1.FullScheduleAndRetention{
				Schedule:  fullScheduleModel,
				Retention: retentionModel,
			}

			awsTierModel := &backuprecoveryv1.AWSTier{
				MoveAfterUnit: core.StringPtr("Days"),
				MoveAfter:     core.Int64Ptr(int64(26)),
				TierType:      core.StringPtr("kAmazonS3Standard"),
			}

			awsTiersModel := &backuprecoveryv1.AWSTiers{
				Tiers: []backuprecoveryv1.AWSTier{*awsTierModel},
			}

			azureTierModel := &backuprecoveryv1.AzureTier{
				MoveAfterUnit: core.StringPtr("Days"),
				MoveAfter:     core.Int64Ptr(int64(26)),
				TierType:      core.StringPtr("kAzureTierHot"),
			}

			azureTiersModel := &backuprecoveryv1.AzureTiers{
				Tiers: []backuprecoveryv1.AzureTier{*azureTierModel},
			}

			googleTierModel := &backuprecoveryv1.GoogleTier{
				MoveAfterUnit: core.StringPtr("Days"),
				MoveAfter:     core.Int64Ptr(int64(26)),
				TierType:      core.StringPtr("kGoogleStandard"),
			}

			googleTiersModel := &backuprecoveryv1.GoogleTiers{
				Tiers: []backuprecoveryv1.GoogleTier{*googleTierModel},
			}

			oracleTierModel := &backuprecoveryv1.OracleTier{
				MoveAfterUnit: core.StringPtr("Days"),
				MoveAfter:     core.Int64Ptr(int64(26)),
				TierType:      core.StringPtr("kOracleTierStandard"),
			}

			oracleTiersModel := &backuprecoveryv1.OracleTiers{
				Tiers: []backuprecoveryv1.OracleTier{*oracleTierModel},
			}

			tierLevelSettingsModel := &backuprecoveryv1.TierLevelSettings{
				AwsTiering:    awsTiersModel,
				AzureTiering:  azureTiersModel,
				CloudPlatform: core.StringPtr("AWS"),
				GoogleTiering: googleTiersModel,
				OracleTiering: oracleTiersModel,
			}

			primaryArchivalTargetModel := &backuprecoveryv1.PrimaryArchivalTarget{
				TargetID:     core.Int64Ptr(int64(26)),
				TierSettings: tierLevelSettingsModel,
			}

			primaryBackupTargetModel := &backuprecoveryv1.PrimaryBackupTarget{
				TargetType:             core.StringPtr("Local"),
				ArchivalTargetSettings: primaryArchivalTargetModel,
				UseDefaultBackupTarget: core.BoolPtr(true),
			}

			regularBackupPolicyModel := &backuprecoveryv1.RegularBackupPolicy{
				Incremental:         incrementalBackupPolicyModel,
				Full:                fullBackupPolicyModel,
				FullBackups:         []backuprecoveryv1.FullScheduleAndRetention{*fullScheduleAndRetentionModel},
				Retention:           retentionModel,
				PrimaryBackupTarget: primaryBackupTargetModel,
			}

			logScheduleModel := &backuprecoveryv1.LogSchedule{
				Unit:           core.StringPtr("Minutes"),
				MinuteSchedule: minuteScheduleModel,
				HourSchedule:   hourScheduleModel,
			}

			logBackupPolicyModel := &backuprecoveryv1.LogBackupPolicy{
				Schedule:  logScheduleModel,
				Retention: retentionModel,
			}

			bmrScheduleModel := &backuprecoveryv1.BmrSchedule{
				Unit:          core.StringPtr("Days"),
				DaySchedule:   dayScheduleModel,
				WeekSchedule:  weekScheduleModel,
				MonthSchedule: monthScheduleModel,
				YearSchedule:  yearScheduleModel,
			}

			bmrBackupPolicyModel := &backuprecoveryv1.BmrBackupPolicy{
				Schedule:  bmrScheduleModel,
				Retention: retentionModel,
			}

			cdpRetentionModel := &backuprecoveryv1.CdpRetention{
				Unit:           core.StringPtr("Minutes"),
				Duration:       core.Int64Ptr(int64(1)),
				DataLockConfig: dataLockConfigModel,
			}

			cdpBackupPolicyModel := &backuprecoveryv1.CdpBackupPolicy{
				Retention: cdpRetentionModel,
			}

			storageArraySnapshotScheduleModel := &backuprecoveryv1.StorageArraySnapshotSchedule{
				Unit:           core.StringPtr("Minutes"),
				MinuteSchedule: minuteScheduleModel,
				HourSchedule:   hourScheduleModel,
				DaySchedule:    dayScheduleModel,
				WeekSchedule:   weekScheduleModel,
				MonthSchedule:  monthScheduleModel,
				YearSchedule:   yearScheduleModel,
			}

			storageArraySnapshotBackupPolicyModel := &backuprecoveryv1.StorageArraySnapshotBackupPolicy{
				Schedule:  storageArraySnapshotScheduleModel,
				Retention: retentionModel,
			}

			cancellationTimeoutParamsModel := &backuprecoveryv1.CancellationTimeoutParams{
				TimeoutMins: core.Int64Ptr(int64(26)),
				BackupType:  core.StringPtr("kRegular"),
			}

			backupPolicyModel := &backuprecoveryv1.BackupPolicy{
				Regular:              regularBackupPolicyModel,
				Log:                  logBackupPolicyModel,
				Bmr:                  bmrBackupPolicyModel,
				Cdp:                  cdpBackupPolicyModel,
				StorageArraySnapshot: storageArraySnapshotBackupPolicyModel,
				RunTimeouts:          []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
			}

			timeOfDayModel := &backuprecoveryv1.TimeOfDay{
				Hour:     core.Int64Ptr(int64(1)),
				Minute:   core.Int64Ptr(int64(15)),
				TimeZone: core.StringPtr("America/Los_Angeles"),
			}

			blackoutWindowModel := &backuprecoveryv1.BlackoutWindow{
				Day:       core.StringPtr("Sunday"),
				StartTime: timeOfDayModel,
				EndTime:   timeOfDayModel,
				ConfigID:  core.StringPtr("Config-Id"),
			}

			extendedRetentionScheduleModel := &backuprecoveryv1.ExtendedRetentionSchedule{
				Unit:      core.StringPtr("Runs"),
				Frequency: core.Int64Ptr(int64(3)),
			}

			extendedRetentionPolicyModel := &backuprecoveryv1.ExtendedRetentionPolicy{
				Schedule:  extendedRetentionScheduleModel,
				Retention: retentionModel,
				RunType:   core.StringPtr("Regular"),
				ConfigID:  core.StringPtr("Config-Id"),
			}

			targetScheduleModel := &backuprecoveryv1.TargetSchedule{
				Unit:      core.StringPtr("Runs"),
				Frequency: core.Int64Ptr(int64(3)),
			}

			logRetentionModel := &backuprecoveryv1.LogRetention{
				Unit:           core.StringPtr("Days"),
				Duration:       core.Int64Ptr(int64(0)),
				DataLockConfig: dataLockConfigModel,
			}

			awsTargetConfigModel := &backuprecoveryv1.AWSTargetConfig{
				Region:   core.Int64Ptr(int64(26)),
				SourceID: core.Int64Ptr(int64(26)),
			}

			azureTargetConfigModel := &backuprecoveryv1.AzureTargetConfig{
				ResourceGroup: core.Int64Ptr(int64(26)),
				SourceID:      core.Int64Ptr(int64(26)),
			}

			remoteTargetConfigModel := &backuprecoveryv1.RemoteTargetConfig{
				ClusterID: core.Int64Ptr(int64(26)),
			}

			replicationTargetConfigurationModel := &backuprecoveryv1.ReplicationTargetConfiguration{
				Schedule:           targetScheduleModel,
				Retention:          retentionModel,
				CopyOnRunSuccess:   core.BoolPtr(true),
				ConfigID:           core.StringPtr("Config-Id"),
				BackupRunType:      core.StringPtr("Regular"),
				RunTimeouts:        []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				LogRetention:       logRetentionModel,
				AwsTargetConfig:    awsTargetConfigModel,
				AzureTargetConfig:  azureTargetConfigModel,
				TargetType:         core.StringPtr("RemoteCluster"),
				RemoteTargetConfig: remoteTargetConfigModel,
			}

			archivalTargetConfigurationModel := &backuprecoveryv1.ArchivalTargetConfiguration{
				Schedule:          targetScheduleModel,
				Retention:         retentionModel,
				CopyOnRunSuccess:  core.BoolPtr(true),
				ConfigID:          core.StringPtr("Config-Id"),
				BackupRunType:     core.StringPtr("Regular"),
				RunTimeouts:       []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				LogRetention:      logRetentionModel,
				TargetID:          core.Int64Ptr(int64(5)),
				TierSettings:      tierLevelSettingsModel,
				ExtendedRetention: []backuprecoveryv1.ExtendedRetentionPolicy{*extendedRetentionPolicyModel},
			}

			customTagParamsModel := &backuprecoveryv1.CustomTagParams{
				Key:   core.StringPtr("custom-tag-key"),
				Value: core.StringPtr("custom-tag-value"),
			}

			awsCloudSpinParamsModel := &backuprecoveryv1.AwsCloudSpinParams{
				CustomTagList: []backuprecoveryv1.CustomTagParams{*customTagParamsModel},
				Region:        core.Int64Ptr(int64(3)),
				SubnetID:      core.Int64Ptr(int64(26)),
				VpcID:         core.Int64Ptr(int64(26)),
			}

			azureCloudSpinParamsModel := &backuprecoveryv1.AzureCloudSpinParams{
				AvailabilitySetID:        core.Int64Ptr(int64(26)),
				NetworkResourceGroupID:   core.Int64Ptr(int64(26)),
				ResourceGroupID:          core.Int64Ptr(int64(26)),
				StorageAccountID:         core.Int64Ptr(int64(26)),
				StorageContainerID:       core.Int64Ptr(int64(26)),
				StorageResourceGroupID:   core.Int64Ptr(int64(26)),
				TempVmResourceGroupID:    core.Int64Ptr(int64(26)),
				TempVmStorageAccountID:   core.Int64Ptr(int64(26)),
				TempVmStorageContainerID: core.Int64Ptr(int64(26)),
				TempVmSubnetID:           core.Int64Ptr(int64(26)),
				TempVmVirtualNetworkID:   core.Int64Ptr(int64(26)),
			}

			cloudSpinTargetModel := &backuprecoveryv1.CloudSpinTarget{
				AwsParams:   awsCloudSpinParamsModel,
				AzureParams: azureCloudSpinParamsModel,
				ID:          core.Int64Ptr(int64(2)),
			}

			cloudSpinTargetConfigurationModel := &backuprecoveryv1.CloudSpinTargetConfiguration{
				Schedule:         targetScheduleModel,
				Retention:        retentionModel,
				CopyOnRunSuccess: core.BoolPtr(true),
				ConfigID:         core.StringPtr("Config-Id"),
				BackupRunType:    core.StringPtr("Regular"),
				RunTimeouts:      []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				LogRetention:     logRetentionModel,
				Target:           cloudSpinTargetModel,
			}

			onpremDeployParamsModel := &backuprecoveryv1.OnpremDeployParams{
				ID: core.Int64Ptr(int64(4)),
			}

			onpremDeployTargetConfigurationModel := &backuprecoveryv1.OnpremDeployTargetConfiguration{
				Schedule:         targetScheduleModel,
				Retention:        retentionModel,
				CopyOnRunSuccess: core.BoolPtr(true),
				ConfigID:         core.StringPtr("Config-Id"),
				BackupRunType:    core.StringPtr("Regular"),
				RunTimeouts:      []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				LogRetention:     logRetentionModel,
				Params:           onpremDeployParamsModel,
			}

			rpaasTargetConfigurationModel := &backuprecoveryv1.RpaasTargetConfiguration{
				Schedule:         targetScheduleModel,
				Retention:        retentionModel,
				CopyOnRunSuccess: core.BoolPtr(true),
				ConfigID:         core.StringPtr("Config-Id"),
				BackupRunType:    core.StringPtr("Regular"),
				RunTimeouts:      []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				LogRetention:     logRetentionModel,
				TargetID:         core.Int64Ptr(int64(5)),
				TargetType:       core.StringPtr("Tape"),
			}

			targetsConfigurationModel := &backuprecoveryv1.TargetsConfiguration{
				ReplicationTargets:  []backuprecoveryv1.ReplicationTargetConfiguration{*replicationTargetConfigurationModel},
				ArchivalTargets:     []backuprecoveryv1.ArchivalTargetConfiguration{*archivalTargetConfigurationModel},
				CloudSpinTargets:    []backuprecoveryv1.CloudSpinTargetConfiguration{*cloudSpinTargetConfigurationModel},
				OnpremDeployTargets: []backuprecoveryv1.OnpremDeployTargetConfiguration{*onpremDeployTargetConfigurationModel},
				RpaasTargets:        []backuprecoveryv1.RpaasTargetConfiguration{*rpaasTargetConfigurationModel},
			}

			cascadedTargetConfigurationModel := &backuprecoveryv1.CascadedTargetConfiguration{
				SourceClusterID: core.Int64Ptr(int64(26)),
				RemoteTargets:   targetsConfigurationModel,
			}

			retryOptionsModel := &backuprecoveryv1.RetryOptions{
				Retries:           core.Int64Ptr(int64(0)),
				RetryIntervalMins: core.Int64Ptr(int64(1)),
			}

			updateProtectionPolicyOptions := &backuprecoveryv1.UpdateProtectionPolicyOptions{
				ID:                        &policyIdLink,
				XIBMTenantID:              core.StringPtr("tenantId"),
				Name:                      core.StringPtr("update-protection-policy"),
				BackupPolicy:              backupPolicyModel,
				Description:               core.StringPtr("Protection Policy"),
				BlackoutWindow:            []backuprecoveryv1.BlackoutWindow{*blackoutWindowModel},
				ExtendedRetention:         []backuprecoveryv1.ExtendedRetentionPolicy{*extendedRetentionPolicyModel},
				RemoteTargetPolicy:        targetsConfigurationModel,
				CascadedTargetsConfig:     []backuprecoveryv1.CascadedTargetConfiguration{*cascadedTargetConfigurationModel},
				RetryOptions:              retryOptionsModel,
				DataLock:                  core.StringPtr("Compliance"),
				Version:                   core.Int64Ptr(int64(38)),
				IsCBSEnabled:              core.BoolPtr(true),
				LastModificationTimeUsecs: core.Int64Ptr(int64(26)),
				TemplateID:                core.StringPtr("protection-policy-template"),
			}

			protectionPolicyResponse, response, err := backupRecoveryService.UpdateProtectionPolicy(updateProtectionPolicyOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(protectionPolicyResponse).ToNot(BeNil())
		})
	})

	Describe(`GetProtectionGroups - Get the list of Protection Groups`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetProtectionGroups(getProtectionGroupsOptions *GetProtectionGroupsOptions)`, func() {
			getProtectionGroupsOptions := &backuprecoveryv1.GetProtectionGroupsOptions{
				XIBMTenantID:                  core.StringPtr("tenantID"),
				RequestInitiatorType:          core.StringPtr("UIUser"),
				Ids:                           []string{"protectionGroupId1"},
				Names:                         []string{"policyName1"},
				PolicyIds:                     []string{"policyId1"},
				IncludeGroupsWithDatalockOnly: core.BoolPtr(true),
				Environments:                  []string{"kPhysical"},
				IsActive:                      core.BoolPtr(true),
				IsDeleted:                     core.BoolPtr(true),
				IsPaused:                      core.BoolPtr(true),
				LastRunLocalBackupStatus:      []string{"Accepted"},
				LastRunReplicationStatus:      []string{"Accepted"},
				LastRunArchivalStatus:         []string{"Accepted"},
				LastRunCloudSpinStatus:        []string{"Accepted"},
				LastRunAnyStatus:              []string{"Accepted"},
				IsLastRunSlaViolated:          core.BoolPtr(true),
				IncludeLastRunInfo:            core.BoolPtr(true),
				PruneExcludedSourceIds:        core.BoolPtr(true),
				PruneSourceIds:                core.BoolPtr(true),
				UseCachedData:                 core.BoolPtr(true),
				SourceIds:                     []int64{int64(26)},
			}

			protectionGroupsResponse, response, err := backupRecoveryService.GetProtectionGroups(getProtectionGroupsOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(protectionGroupsResponse).ToNot(BeNil())
		})
	})

	Describe(`GetProtectionGroupByID - List details about single Protection Group`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetProtectionGroupByID(getProtectionGroupByIdOptions *GetProtectionGroupByIdOptions)`, func() {
			getProtectionGroupByIdOptions := &backuprecoveryv1.GetProtectionGroupByIdOptions{
				ID:                     &protectionGroupIdLink,
				XIBMTenantID:           core.StringPtr("tenantID"),
				RequestInitiatorType:   core.StringPtr("UIUser"),
				IncludeLastRunInfo:     core.BoolPtr(true),
				PruneExcludedSourceIds: core.BoolPtr(true),
				PruneSourceIds:         core.BoolPtr(true),
			}

			protectionGroupResponse, response, err := backupRecoveryService.GetProtectionGroupByID(getProtectionGroupByIdOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(protectionGroupResponse).ToNot(BeNil())
		})
	})

	Describe(`UpdateProtectionGroup - Update a Protection Group`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`UpdateProtectionGroup(updateProtectionGroupOptions *UpdateProtectionGroupOptions)`, func() {
			timeOfDayModel := &backuprecoveryv1.TimeOfDay{
				Hour:     core.Int64Ptr(int64(0)),
				Minute:   core.Int64Ptr(int64(0)),
				TimeZone: core.StringPtr("America/Los_Angeles"),
			}

			alertTargetModel := &backuprecoveryv1.AlertTarget{
				EmailAddress:  core.StringPtr("alert1@domain.com"),
				Language:      core.StringPtr("en-us"),
				RecipientType: core.StringPtr("kTo"),
			}

			protectionGroupAlertingPolicyModel := &backuprecoveryv1.ProtectionGroupAlertingPolicy{
				BackupRunStatus:              []string{"kSuccess"},
				AlertTargets:                 []backuprecoveryv1.AlertTarget{*alertTargetModel},
				RaiseObjectLevelFailureAlert: core.BoolPtr(true),
				RaiseObjectLevelFailureAlertAfterLastAttempt: core.BoolPtr(true),
				RaiseObjectLevelFailureAlertAfterEachAttempt: core.BoolPtr(true),
			}

			slaRuleModel := &backuprecoveryv1.SlaRule{
				BackupRunType: core.StringPtr("kIncremental"),
				SlaMinutes:    core.Int64Ptr(int64(1)),
			}

			keyValuePairModel := &backuprecoveryv1.KeyValuePair{
				Key:   core.StringPtr("configKey"),
				Value: core.StringPtr("configValue"),
			}

			physicalVolumeProtectionGroupObjectParamsModel := &backuprecoveryv1.PhysicalVolumeProtectionGroupObjectParams{
				ID:                 core.Int64Ptr(int64(3)),
				VolumeGuids:        []string{"volumeGuid1"},
				EnableSystemBackup: core.BoolPtr(true),
				ExcludedVssWriters: []string{"writerName1", "writerName2"},
			}

			indexingPolicyModel := &backuprecoveryv1.IndexingPolicy{
				EnableIndexing: core.BoolPtr(true),
				IncludePaths:   []string{"~/dir1"},
				ExcludePaths:   []string{"~/dir2"},
			}

			commonPreBackupScriptParamsModel := &backuprecoveryv1.CommonPreBackupScriptParams{
				Path:            core.StringPtr("~/script1"),
				Params:          core.StringPtr("param1"),
				TimeoutSecs:     core.Int64Ptr(int64(1)),
				IsActive:        core.BoolPtr(true),
				ContinueOnError: core.BoolPtr(true),
			}

			commonPostBackupScriptParamsModel := &backuprecoveryv1.CommonPostBackupScriptParams{
				Path:        core.StringPtr("~/script2"),
				Params:      core.StringPtr("param2"),
				TimeoutSecs: core.Int64Ptr(int64(1)),
				IsActive:    core.BoolPtr(true),
			}

			prePostScriptParamsModel := &backuprecoveryv1.PrePostScriptParams{
				PreScript:  commonPreBackupScriptParamsModel,
				PostScript: commonPostBackupScriptParamsModel,
			}

			physicalVolumeProtectionGroupParamsModel := &backuprecoveryv1.PhysicalVolumeProtectionGroupParams{
				Objects:                        []backuprecoveryv1.PhysicalVolumeProtectionGroupObjectParams{*physicalVolumeProtectionGroupObjectParamsModel},
				IndexingPolicy:                 indexingPolicyModel,
				PerformSourceSideDeduplication: core.BoolPtr(true),
				Quiesce:                        core.BoolPtr(true),
				ContinueOnQuiesceFailure:       core.BoolPtr(true),
				IncrementalBackupAfterRestart:  core.BoolPtr(true),
				PrePostScript:                  prePostScriptParamsModel,
				DedupExclusionSourceIds:        []int64{int64(26)},
				ExcludedVssWriters:             []string{"writerName1", "writerName2"},
				CobmrBackup:                    core.BoolPtr(true),
			}

			physicalFileBackupPathParamsModel := &backuprecoveryv1.PhysicalFileBackupPathParams{
				IncludedPath:      core.StringPtr("~/dir1/"),
				ExcludedPaths:     []string{"~/dir2"},
				SkipNestedVolumes: core.BoolPtr(true),
			}

			physicalFileProtectionGroupObjectParamsModel := &backuprecoveryv1.PhysicalFileProtectionGroupObjectParams{
				ExcludedVssWriters:                   []string{"writerName1", "writerName2"},
				ID:                                   core.Int64Ptr(int64(2)),
				FilePaths:                            []backuprecoveryv1.PhysicalFileBackupPathParams{*physicalFileBackupPathParamsModel},
				UsesPathLevelSkipNestedVolumeSetting: core.BoolPtr(true),
				NestedVolumeTypesToSkip:              []string{"volume1"},
				FollowNasSymlinkTarget:               core.BoolPtr(true),
				MetadataFilePath:                     core.StringPtr("~/dir3"),
			}

			cancellationTimeoutParamsModel := &backuprecoveryv1.CancellationTimeoutParams{
				TimeoutMins: core.Int64Ptr(int64(26)),
				BackupType:  core.StringPtr("kRegular"),
			}

			physicalFileProtectionGroupParamsModel := &backuprecoveryv1.PhysicalFileProtectionGroupParams{
				ExcludedVssWriters:             []string{"writerName1", "writerName2"},
				Objects:                        []backuprecoveryv1.PhysicalFileProtectionGroupObjectParams{*physicalFileProtectionGroupObjectParamsModel},
				IndexingPolicy:                 indexingPolicyModel,
				PerformSourceSideDeduplication: core.BoolPtr(true),
				PerformBrickBasedDeduplication: core.BoolPtr(true),
				TaskTimeouts:                   []backuprecoveryv1.CancellationTimeoutParams{*cancellationTimeoutParamsModel},
				Quiesce:                        core.BoolPtr(true),
				ContinueOnQuiesceFailure:       core.BoolPtr(true),
				CobmrBackup:                    core.BoolPtr(true),
				PrePostScript:                  prePostScriptParamsModel,
				DedupExclusionSourceIds:        []int64{int64(26)},
				GlobalExcludePaths:             []string{"~/dir1"},
				GlobalExcludeFS:                []string{"~/dir2"},
				IgnorableErrors:                []string{"kEOF"},
				AllowParallelRuns:              core.BoolPtr(true),
			}

			physicalProtectionGroupParamsModel := &backuprecoveryv1.PhysicalProtectionGroupParams{
				ProtectionType:             core.StringPtr("kFile"),
				VolumeProtectionTypeParams: physicalVolumeProtectionGroupParamsModel,
				FileProtectionTypeParams:   physicalFileProtectionGroupParamsModel,
			}

			advancedSettingsModel := &backuprecoveryv1.AdvancedSettings{
				ClonedDbBackupStatus:            core.StringPtr("kError"),
				DbBackupIfNotOnlineStatus:       core.StringPtr("kError"),
				MissingDbBackupStatus:           core.StringPtr("kError"),
				OfflineRestoringDbBackupStatus:  core.StringPtr("kError"),
				ReadOnlyDbBackupStatus:          core.StringPtr("kError"),
				ReportAllNonAutoprotectDbErrors: core.StringPtr("kError"),
			}

			filterModel := &backuprecoveryv1.Filter{
				FilterString:        core.StringPtr("filterString"),
				IsRegularExpression: core.BoolPtr(false),
			}

			mssqlFileProtectionGroupHostParamsModel := &backuprecoveryv1.MSSQLFileProtectionGroupHostParams{
				DisableSourceSideDeduplication: core.BoolPtr(true),
				HostID:                         core.Int64Ptr(int64(26)),
			}

			mssqlFileProtectionGroupObjectParamsModel := &backuprecoveryv1.MSSQLFileProtectionGroupObjectParams{
				ID: core.Int64Ptr(int64(6)),
			}

			mssqlFileProtectionGroupParamsModel := &backuprecoveryv1.MSSQLFileProtectionGroupParams{
				AagBackupPreferenceType:        core.StringPtr("kPrimaryReplicaOnly"),
				AdvancedSettings:               advancedSettingsModel,
				BackupSystemDbs:                core.BoolPtr(true),
				ExcludeFilters:                 []backuprecoveryv1.Filter{*filterModel},
				FullBackupsCopyOnly:            core.BoolPtr(true),
				LogBackupNumStreams:            core.Int64Ptr(int64(38)),
				LogBackupWithClause:            core.StringPtr("backupWithClause"),
				PrePostScript:                  prePostScriptParamsModel,
				UseAagPreferencesFromServer:    core.BoolPtr(true),
				UserDbBackupPreferenceType:     core.StringPtr("kBackupAllDatabases"),
				AdditionalHostParams:           []backuprecoveryv1.MSSQLFileProtectionGroupHostParams{*mssqlFileProtectionGroupHostParamsModel},
				Objects:                        []backuprecoveryv1.MSSQLFileProtectionGroupObjectParams{*mssqlFileProtectionGroupObjectParamsModel},
				PerformSourceSideDeduplication: core.BoolPtr(true),
			}

			mssqlNativeProtectionGroupObjectParamsModel := &backuprecoveryv1.MSSQLNativeProtectionGroupObjectParams{
				ID: core.Int64Ptr(int64(6)),
			}

			mssqlNativeProtectionGroupParamsModel := &backuprecoveryv1.MSSQLNativeProtectionGroupParams{
				AagBackupPreferenceType:     core.StringPtr("kPrimaryReplicaOnly"),
				AdvancedSettings:            advancedSettingsModel,
				BackupSystemDbs:             core.BoolPtr(true),
				ExcludeFilters:              []backuprecoveryv1.Filter{*filterModel},
				FullBackupsCopyOnly:         core.BoolPtr(true),
				LogBackupNumStreams:         core.Int64Ptr(int64(38)),
				LogBackupWithClause:         core.StringPtr("backupWithClause"),
				PrePostScript:               prePostScriptParamsModel,
				UseAagPreferencesFromServer: core.BoolPtr(true),
				UserDbBackupPreferenceType:  core.StringPtr("kBackupAllDatabases"),
				NumStreams:                  core.Int64Ptr(int64(38)),
				Objects:                     []backuprecoveryv1.MSSQLNativeProtectionGroupObjectParams{*mssqlNativeProtectionGroupObjectParamsModel},
				WithClause:                  core.StringPtr("withClause"),
			}

			mssqlVolumeProtectionGroupHostParamsModel := &backuprecoveryv1.MSSQLVolumeProtectionGroupHostParams{
				EnableSystemBackup: core.BoolPtr(true),
				HostID:             core.Int64Ptr(int64(8)),
				VolumeGuids:        []string{"volumeGuid1"},
			}

			mssqlVolumeProtectionGroupObjectParamsModel := &backuprecoveryv1.MSSQLVolumeProtectionGroupObjectParams{
				ID: core.Int64Ptr(int64(6)),
			}

			mssqlVolumeProtectionGroupParamsModel := &backuprecoveryv1.MSSQLVolumeProtectionGroupParams{
				AagBackupPreferenceType:       core.StringPtr("kPrimaryReplicaOnly"),
				AdvancedSettings:              advancedSettingsModel,
				BackupSystemDbs:               core.BoolPtr(true),
				ExcludeFilters:                []backuprecoveryv1.Filter{*filterModel},
				FullBackupsCopyOnly:           core.BoolPtr(true),
				LogBackupNumStreams:           core.Int64Ptr(int64(38)),
				LogBackupWithClause:           core.StringPtr("backupWithClause"),
				PrePostScript:                 prePostScriptParamsModel,
				UseAagPreferencesFromServer:   core.BoolPtr(true),
				UserDbBackupPreferenceType:    core.StringPtr("kBackupAllDatabases"),
				AdditionalHostParams:          []backuprecoveryv1.MSSQLVolumeProtectionGroupHostParams{*mssqlVolumeProtectionGroupHostParamsModel},
				BackupDbVolumesOnly:           core.BoolPtr(true),
				IncrementalBackupAfterRestart: core.BoolPtr(true),
				IndexingPolicy:                indexingPolicyModel,
				Objects:                       []backuprecoveryv1.MSSQLVolumeProtectionGroupObjectParams{*mssqlVolumeProtectionGroupObjectParamsModel},
			}

			mssqlProtectionGroupParamsModel := &backuprecoveryv1.MSSQLProtectionGroupParams{
				FileProtectionTypeParams:   mssqlFileProtectionGroupParamsModel,
				NativeProtectionTypeParams: mssqlNativeProtectionGroupParamsModel,
				ProtectionType:             core.StringPtr("kFile"),
				VolumeProtectionTypeParams: mssqlVolumeProtectionGroupParamsModel,
			}

			updateProtectionGroupOptions := &backuprecoveryv1.UpdateProtectionGroupOptions{
				ID:                         &protectionGroupIdLink,
				XIBMTenantID:               core.StringPtr("tenantId"),
				Name:                       core.StringPtr("update-protection-group"),
				PolicyID:                   core.StringPtr("xxxxxxxxxxxxxxxx:xxxxxxxxxxxxx:xx"),
				Environment:                core.StringPtr("kPhysical"),
				Priority:                   core.StringPtr("kLow"),
				Description:                core.StringPtr("Protection Group"),
				StartTime:                  timeOfDayModel,
				EndTimeUsecs:               core.Int64Ptr(int64(26)),
				LastModifiedTimestampUsecs: core.Int64Ptr(int64(26)),
				AlertPolicy:                protectionGroupAlertingPolicyModel,
				Sla:                        []backuprecoveryv1.SlaRule{*slaRuleModel},
				QosPolicy:                  core.StringPtr("kBackupHDD"),
				AbortInBlackouts:           core.BoolPtr(true),
				PauseInBlackouts:           core.BoolPtr(true),
				IsPaused:                   core.BoolPtr(true),
				AdvancedConfigs:            []backuprecoveryv1.KeyValuePair{*keyValuePairModel},
				PhysicalParams:             physicalProtectionGroupParamsModel,
				MssqlParams:                mssqlProtectionGroupParamsModel,
			}

			protectionGroupResponse, response, err := backupRecoveryService.UpdateProtectionGroup(updateProtectionGroupOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(protectionGroupResponse).ToNot(BeNil())
		})
	})

	Describe(`GetProtectionGroupRuns - Get the list of runs for a Protection Group`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetProtectionGroupRuns(getProtectionGroupRunsOptions *GetProtectionGroupRunsOptions)`, func() {
			getProtectionGroupRunsOptions := &backuprecoveryv1.GetProtectionGroupRunsOptions{
				ID:                          &protectionGroupIdLink,
				XIBMTenantID:                core.StringPtr("tenantId"),
				RequestInitiatorType:        core.StringPtr("UIUser"),
				RunID:                       core.StringPtr("11:111"),
				StartTimeUsecs:              core.Int64Ptr(int64(26)),
				EndTimeUsecs:                core.Int64Ptr(int64(26)),
				RunTypes:                    []string{"kAll"},
				IncludeObjectDetails:        core.BoolPtr(true),
				LocalBackupRunStatus:        []string{"Accepted"},
				ReplicationRunStatus:        []string{"Accepted"},
				ArchivalRunStatus:           []string{"Accepted"},
				CloudSpinRunStatus:          []string{"Accepted"},
				NumRuns:                     core.Int64Ptr(int64(26)),
				ExcludeNonRestorableRuns:    core.BoolPtr(false),
				RunTags:                     []string{"tag1"},
				UseCachedData:               core.BoolPtr(true),
				FilterByEndTime:             core.BoolPtr(true),
				SnapshotTargetTypes:         []string{"Local"},
				OnlyReturnSuccessfulCopyRun: core.BoolPtr(true),
				FilterByCopyTaskEndTime:     core.BoolPtr(true),
			}

			protectionGroupRunsResponse, response, err := backupRecoveryService.GetProtectionGroupRuns(getProtectionGroupRunsOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(protectionGroupRunsResponse).ToNot(BeNil())
		})
	})

	Describe(`UpdateProtectionGroupRun - Update runs for a particular Protection Group`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`UpdateProtectionGroupRun(updateProtectionGroupRunOptions *UpdateProtectionGroupRunOptions)`, func() {
			updateLocalSnapshotConfigModel := &backuprecoveryv1.UpdateLocalSnapshotConfig{
				EnableLegalHold: core.BoolPtr(true),
				DeleteSnapshot:  core.BoolPtr(true),
				DataLock:        core.StringPtr("Compliance"),
				DaysToKeep:      core.Int64Ptr(int64(26)),
			}

			dataLockConfigModel := &backuprecoveryv1.DataLockConfig{
				Mode:                       core.StringPtr("Compliance"),
				Unit:                       core.StringPtr("Days"),
				Duration:                   core.Int64Ptr(int64(1)),
				EnableWormOnExternalTarget: core.BoolPtr(true),
			}

			retentionModel := &backuprecoveryv1.Retention{
				Unit:           core.StringPtr("Days"),
				Duration:       core.Int64Ptr(int64(1)),
				DataLockConfig: dataLockConfigModel,
			}

			runReplicationConfigModel := &backuprecoveryv1.RunReplicationConfig{
				ID:        core.Int64Ptr(int64(26)),
				Retention: retentionModel,
			}

			updateExistingReplicationSnapshotConfigModel := &backuprecoveryv1.UpdateExistingReplicationSnapshotConfig{
				ID:              core.Int64Ptr(int64(4)),
				Name:            core.StringPtr("update-snapshot-config"),
				EnableLegalHold: core.BoolPtr(true),
				DeleteSnapshot:  core.BoolPtr(true),
				Resync:          core.BoolPtr(true),
				DataLock:        core.StringPtr("Compliance"),
				DaysToKeep:      core.Int64Ptr(int64(26)),
			}

			updateReplicationSnapshotConfigModel := &backuprecoveryv1.UpdateReplicationSnapshotConfig{
				NewSnapshotConfig:            []backuprecoveryv1.RunReplicationConfig{*runReplicationConfigModel},
				UpdateExistingSnapshotConfig: []backuprecoveryv1.UpdateExistingReplicationSnapshotConfig{*updateExistingReplicationSnapshotConfigModel},
			}

			runArchivalConfigModel := &backuprecoveryv1.RunArchivalConfig{
				ID:                      core.Int64Ptr(int64(2)),
				ArchivalTargetType:      core.StringPtr("Tape"),
				Retention:               retentionModel,
				CopyOnlyFullySuccessful: core.BoolPtr(true),
			}

			updateExistingArchivalSnapshotConfigModel := &backuprecoveryv1.UpdateExistingArchivalSnapshotConfig{
				ID:                 core.Int64Ptr(int64(3)),
				Name:               core.StringPtr("update-snapshot-config"),
				ArchivalTargetType: core.StringPtr("Tape"),
				EnableLegalHold:    core.BoolPtr(true),
				DeleteSnapshot:     core.BoolPtr(true),
				Resync:             core.BoolPtr(true),
				DataLock:           core.StringPtr("Compliance"),
				DaysToKeep:         core.Int64Ptr(int64(26)),
			}

			updateArchivalSnapshotConfigModel := &backuprecoveryv1.UpdateArchivalSnapshotConfig{
				NewSnapshotConfig:            []backuprecoveryv1.RunArchivalConfig{*runArchivalConfigModel},
				UpdateExistingSnapshotConfig: []backuprecoveryv1.UpdateExistingArchivalSnapshotConfig{*updateExistingArchivalSnapshotConfigModel},
			}

			updateProtectionGroupRunParamsModel := &backuprecoveryv1.UpdateProtectionGroupRunParams{
				RunID:                     core.StringPtr("11:111"),
				LocalSnapshotConfig:       updateLocalSnapshotConfigModel,
				ReplicationSnapshotConfig: updateReplicationSnapshotConfigModel,
				ArchivalSnapshotConfig:    updateArchivalSnapshotConfigModel,
			}

			updateProtectionGroupRunOptions := &backuprecoveryv1.UpdateProtectionGroupRunOptions{
				ID:                             &protectionGroupIdLink,
				XIBMTenantID:                   core.StringPtr("tenantId"),
				UpdateProtectionGroupRunParams: []backuprecoveryv1.UpdateProtectionGroupRunParams{*updateProtectionGroupRunParamsModel},
			}

			updateProtectionGroupRunResponse, response, err := backupRecoveryService.UpdateProtectionGroupRun(updateProtectionGroupRunOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(207))
			Expect(updateProtectionGroupRunResponse).ToNot(BeNil())
		})
	})

	Describe(`CreateProtectionGroupRun - Create a new protection run`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`CreateProtectionGroupRun(createProtectionGroupRunOptions *CreateProtectionGroupRunOptions)`, func() {
			runObjectPhysicalParamsModel := &backuprecoveryv1.RunObjectPhysicalParams{
				MetadataFilePath: core.StringPtr("~/metadata"),
			}

			runObjectModel := &backuprecoveryv1.RunObject{
				ID:             core.Int64Ptr(int64(4)),
				AppIds:         []int64{int64(26)},
				PhysicalParams: runObjectPhysicalParamsModel,
			}

			dataLockConfigModel := &backuprecoveryv1.DataLockConfig{
				Mode:                       core.StringPtr("Compliance"),
				Unit:                       core.StringPtr("Days"),
				Duration:                   core.Int64Ptr(int64(1)),
				EnableWormOnExternalTarget: core.BoolPtr(true),
			}

			retentionModel := &backuprecoveryv1.Retention{
				Unit:           core.StringPtr("Days"),
				Duration:       core.Int64Ptr(int64(1)),
				DataLockConfig: dataLockConfigModel,
			}

			runReplicationConfigModel := &backuprecoveryv1.RunReplicationConfig{
				ID:        core.Int64Ptr(int64(26)),
				Retention: retentionModel,
			}

			runArchivalConfigModel := &backuprecoveryv1.RunArchivalConfig{
				ID:                      core.Int64Ptr(int64(26)),
				ArchivalTargetType:      core.StringPtr("Tape"),
				Retention:               retentionModel,
				CopyOnlyFullySuccessful: core.BoolPtr(true),
			}

			awsTargetConfigModel := &backuprecoveryv1.AWSTargetConfig{
				Region:   core.Int64Ptr(int64(26)),
				SourceID: core.Int64Ptr(int64(26)),
			}

			azureTargetConfigModel := &backuprecoveryv1.AzureTargetConfig{
				ResourceGroup: core.Int64Ptr(int64(26)),
				SourceID:      core.Int64Ptr(int64(26)),
			}

			runCloudReplicationConfigModel := &backuprecoveryv1.RunCloudReplicationConfig{
				AwsTarget:   awsTargetConfigModel,
				AzureTarget: azureTargetConfigModel,
				TargetType:  core.StringPtr("AWS"),
				Retention:   retentionModel,
			}

			runTargetsConfigurationModel := &backuprecoveryv1.RunTargetsConfiguration{
				UsePolicyDefaults: core.BoolPtr(false),
				Replications:      []backuprecoveryv1.RunReplicationConfig{*runReplicationConfigModel},
				Archivals:         []backuprecoveryv1.RunArchivalConfig{*runArchivalConfigModel},
				CloudReplications: []backuprecoveryv1.RunCloudReplicationConfig{*runCloudReplicationConfigModel},
			}

			createProtectionGroupRunOptions := &backuprecoveryv1.CreateProtectionGroupRunOptions{
				ID:            core.StringPtr("runId"),
				XIBMTenantID:  core.StringPtr("tenantId"),
				RunType:       core.StringPtr("kRegular"),
				Objects:       []backuprecoveryv1.RunObject{*runObjectModel},
				TargetsConfig: runTargetsConfigurationModel,
			}

			createProtectionGroupRunResponse, response, err := backupRecoveryService.CreateProtectionGroupRun(createProtectionGroupRunOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(202))
			Expect(createProtectionGroupRunResponse).ToNot(BeNil())
		})
	})

	Describe(`PerformActionOnProtectionGroupRun - Actions on protection group run`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`PerformActionOnProtectionGroupRun(performActionOnProtectionGroupRunOptions *PerformActionOnProtectionGroupRunOptions)`, func() {
			pauseProtectionRunActionParamsModel := &backuprecoveryv1.PauseProtectionRunActionParams{
				RunID: core.StringPtr("11:111"),
			}

			resumeProtectionRunActionParamsModel := &backuprecoveryv1.ResumeProtectionRunActionParams{
				RunID: core.StringPtr("11:111"),
			}

			cancelProtectionGroupRunRequestModel := &backuprecoveryv1.CancelProtectionGroupRunRequest{
				RunID:             core.StringPtr("11:111"),
				LocalTaskID:       core.StringPtr("123:456:789"),
				ObjectIds:         []int64{int64(26)},
				ReplicationTaskID: []string{"123:456:789"},
				ArchivalTaskID:    []string{"123:456:789"},
				CloudSpinTaskID:   []string{"123:456:789"},
			}

			performActionOnProtectionGroupRunOptions := &backuprecoveryv1.PerformActionOnProtectionGroupRunOptions{
				ID:           &protectionGroupIdLink,
				XIBMTenantID: core.StringPtr("tenantId"),
				Action:       core.StringPtr("Pause"),
				PauseParams:  []backuprecoveryv1.PauseProtectionRunActionParams{*pauseProtectionRunActionParamsModel},
				ResumeParams: []backuprecoveryv1.ResumeProtectionRunActionParams{*resumeProtectionRunActionParamsModel},
				CancelParams: []backuprecoveryv1.CancelProtectionGroupRunRequest{*cancelProtectionGroupRunRequestModel},
			}

			performRunActionResponse, response, err := backupRecoveryService.PerformActionOnProtectionGroupRun(performActionOnProtectionGroupRunOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(202))
			Expect(performRunActionResponse).ToNot(BeNil())
		})
	})

	Describe(`GetRecoveries - Lists the Recoveries`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetRecoveries(getRecoveriesOptions *GetRecoveriesOptions)`, func() {
			getRecoveriesOptions := &backuprecoveryv1.GetRecoveriesOptions{
				XIBMTenantID:              core.StringPtr("tenantId"),
				Ids:                       []string{"11:111:11"},
				ReturnOnlyChildRecoveries: core.BoolPtr(true),
				StartTimeUsecs:            core.Int64Ptr(int64(26)),
				EndTimeUsecs:              core.Int64Ptr(int64(26)),
				SnapshotTargetType:        []string{"Local"},
				ArchivalTargetType:        []string{"Tape"},
				SnapshotEnvironments:      []string{"kPhysical"},
				Status:                    []string{"Accepted"},
				RecoveryActions:           []string{"RecoverVMs"},
			}

			recoveriesResponse, response, err := backupRecoveryService.GetRecoveries(getRecoveriesOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(recoveriesResponse).ToNot(BeNil())
		})
	})

	Describe(`CreateDownloadFilesAndFoldersRecovery - Create a download files and folders recovery`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`CreateDownloadFilesAndFoldersRecovery(createDownloadFilesAndFoldersRecoveryOptions *CreateDownloadFilesAndFoldersRecoveryOptions)`, func() {
			commonRecoverObjectSnapshotParamsModel := &backuprecoveryv1.CommonRecoverObjectSnapshotParams{
				SnapshotID:          core.StringPtr("snapshotId"),
				PointInTimeUsecs:    core.Int64Ptr(int64(26)),
				ProtectionGroupID:   core.StringPtr("protectionGroupId"),
				ProtectionGroupName: core.StringPtr("protectionGroupName"),
				RecoverFromStandby:  core.BoolPtr(true),
			}

			filesAndFoldersObjectModel := &backuprecoveryv1.FilesAndFoldersObject{
				AbsolutePath: core.StringPtr("~/home/dir1"),
				IsDirectory:  core.BoolPtr(true),
			}

			documentObjectModel := &backuprecoveryv1.DocumentObject{
				IsDirectory: core.BoolPtr(true),
				ItemID:      core.StringPtr("item1"),
			}

			createDownloadFilesAndFoldersRecoveryOptions := &backuprecoveryv1.CreateDownloadFilesAndFoldersRecoveryOptions{
				XIBMTenantID:         core.StringPtr("tenantId"),
				Name:                 core.StringPtr("create-download-files-and-folders-recovery"),
				Object:               commonRecoverObjectSnapshotParamsModel,
				FilesAndFolders:      []backuprecoveryv1.FilesAndFoldersObject{*filesAndFoldersObjectModel},
				Documents:            []backuprecoveryv1.DocumentObject{*documentObjectModel},
				ParentRecoveryID:     core.StringPtr("parentRecoveryId"),
				GlacierRetrievalType: core.StringPtr("kStandard"),
			}

			recovery, response, err := backupRecoveryService.CreateDownloadFilesAndFoldersRecovery(createDownloadFilesAndFoldersRecoveryOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(201))
			Expect(recovery).ToNot(BeNil())
		})
	})

	Describe(`GetRecoveryByID - Get Recovery for a given id`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetRecoveryByID(getRecoveryByIdOptions *GetRecoveryByIdOptions)`, func() {
			getRecoveryByIdOptions := &backuprecoveryv1.GetRecoveryByIdOptions{
				ID:           &recoveryIdLink,
				XIBMTenantID: core.StringPtr("tenantId"),
			}

			recovery, response, err := backupRecoveryService.GetRecoveryByID(getRecoveryByIdOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(recovery).ToNot(BeNil())
		})
	})

	Describe(`DownloadFilesFromRecovery - Download files from the given download file recovery`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`DownloadFilesFromRecovery(downloadFilesFromRecoveryOptions *DownloadFilesFromRecoveryOptions)`, func() {
			downloadFilesFromRecoveryOptions := &backuprecoveryv1.DownloadFilesFromRecoveryOptions{
				ID:             &recoveryIdLink,
				XIBMTenantID:   core.StringPtr("tenantId"),
				StartOffset:    core.Int64Ptr(int64(26)),
				Length:         core.Int64Ptr(int64(26)),
				FileType:       core.StringPtr("fileType"),
				SourceName:     core.StringPtr("sourceName"),
				StartTime:      core.StringPtr("startTime"),
				IncludeTenants: core.BoolPtr(true),
			}

			response, err := backupRecoveryService.DownloadFilesFromRecovery(downloadFilesFromRecoveryOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
		})
	})

	Describe(`GetRestorePointsInTimeRange - List Restore Points in a given time range`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetRestorePointsInTimeRange(getRestorePointsInTimeRangeOptions *GetRestorePointsInTimeRangeOptions)`, func() {
			getRestorePointsInTimeRangeOptions := &backuprecoveryv1.GetRestorePointsInTimeRangeOptions{
				XIBMTenantID:       core.StringPtr("tenantId"),
				EndTimeUsecs:       core.Int64Ptr(int64(45)),
				Environment:        core.StringPtr("kVMware"),
				ProtectionGroupIds: []string{"protectionGroupId1"},
				StartTimeUsecs:     core.Int64Ptr(int64(15)),
				SourceID:           core.Int64Ptr(int64(26)),
			}

			getRestorePointsInTimeRangeResponse, response, err := backupRecoveryService.GetRestorePointsInTimeRange(getRestorePointsInTimeRangeOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(201))
			Expect(getRestorePointsInTimeRangeResponse).ToNot(BeNil())
		})
	})

	Describe(`DownloadIndexedFile - Download an indexed file`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`DownloadIndexedFile(downloadIndexedFileOptions *DownloadIndexedFileOptions)`, func() {
			downloadIndexedFileOptions := &backuprecoveryv1.DownloadIndexedFileOptions{
				SnapshotsID:  core.StringPtr("snapshotId1"),
				XIBMTenantID: core.StringPtr("tenantId"),
				FilePath:     core.StringPtr("~/home/downloadFile"),
				NvramFile:    core.BoolPtr(true),
				RetryAttempt: core.Int64Ptr(int64(26)),
				StartOffset:  core.Int64Ptr(int64(26)),
				Length:       core.Int64Ptr(int64(26)),
			}

			response, err := backupRecoveryService.DownloadIndexedFile(downloadIndexedFileOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
		})
	})

	Describe(`SearchIndexedObjects - List indexed objects`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`SearchIndexedObjects(searchIndexedObjectsOptions *SearchIndexedObjectsOptions)`, func() {
			cassandraOnPremSearchParamsModel := &backuprecoveryv1.CassandraOnPremSearchParams{
				CassandraObjectTypes: []string{"CassandraKeyspaces"},
				SearchString:         core.StringPtr("searchString"),
				SourceIds:            []int64{int64(26)},
			}

			couchBaseOnPremSearchParamsModel := &backuprecoveryv1.CouchBaseOnPremSearchParams{
				CouchbaseObjectTypes: []string{"CouchbaseBuckets"},
				SearchString:         core.StringPtr("searchString"),
				SourceIds:            []int64{int64(26)},
			}

			o365SearchEmailsRequestParamsModel := &backuprecoveryv1.O365SearchEmailsRequestParams{
				DomainIds:  []int64{int64(26)},
				MailboxIds: []int64{int64(26)},
			}

			searchEmailRequestParamsModel := &backuprecoveryv1.SearchEmailRequestParams{
				AttendeesAddresses:        []string{"attendee1@domain.com"},
				BccRecipientAddresses:     []string{"bccrecipient@domain.com"},
				CcRecipientAddresses:      []string{"ccrecipient@domain.com"},
				CreatedEndTimeSecs:        core.Int64Ptr(int64(26)),
				CreatedStartTimeSecs:      core.Int64Ptr(int64(26)),
				DueDateEndTimeSecs:        core.Int64Ptr(int64(26)),
				DueDateStartTimeSecs:      core.Int64Ptr(int64(26)),
				EmailAddress:              core.StringPtr("email@domain.com"),
				EmailSubject:              core.StringPtr("Email Subject"),
				FirstName:                 core.StringPtr("First Name"),
				FolderNames:               []string{"folder1"},
				HasAttachment:             core.BoolPtr(true),
				LastModifiedEndTimeSecs:   core.Int64Ptr(int64(26)),
				LastModifiedStartTimeSecs: core.Int64Ptr(int64(26)),
				LastName:                  core.StringPtr("Last Name"),
				MiddleName:                core.StringPtr("Middle Name"),
				OrganizerAddress:          core.StringPtr("organizer@domain.com"),
				ReceivedEndTimeSecs:       core.Int64Ptr(int64(26)),
				ReceivedStartTimeSecs:     core.Int64Ptr(int64(26)),
				RecipientAddresses:        []string{"recipient@domain.com"},
				SenderAddress:             core.StringPtr("sender@domain.com"),
				SourceEnvironment:         core.StringPtr("kO365"),
				TaskStatusTypes:           []string{"NotStarted"},
				Types:                     []string{"Email"},
				O365Params:                o365SearchEmailsRequestParamsModel,
			}

			searchExchangeObjectsRequestParamsModel := &backuprecoveryv1.SearchExchangeObjectsRequestParams{
				SearchString: core.StringPtr("searchString"),
			}

			searchFileRequestParamsModel := &backuprecoveryv1.SearchFileRequestParams{
				SearchString:       core.StringPtr("searchString"),
				Types:              []string{"File"},
				SourceEnvironments: []string{"kVMware"},
				SourceIds:          []int64{int64(26)},
				ObjectIds:          []int64{int64(26)},
			}

			hbaseOnPremSearchParamsModel := &backuprecoveryv1.HbaseOnPremSearchParams{
				HbaseObjectTypes: []string{"HbaseNamespaces"},
				SearchString:     core.StringPtr("searchString"),
				SourceIds:        []int64{int64(26)},
			}

			hdfsOnPremSearchParamsModel := &backuprecoveryv1.HDFSOnPremSearchParams{
				HdfsTypes:    []string{"HDFSFolders"},
				SearchString: core.StringPtr("searchString"),
				SourceIds:    []int64{int64(26)},
			}

			hiveOnPremSearchParamsModel := &backuprecoveryv1.HiveOnPremSearchParams{
				HiveObjectTypes: []string{"HiveDatabases"},
				SearchString:    core.StringPtr("searchString"),
				SourceIds:       []int64{int64(26)},
			}

			mongoDbOnPremSearchParamsModel := &backuprecoveryv1.MongoDbOnPremSearchParams{
				MongoDBObjectTypes: []string{"MongoDatabases"},
				SearchString:       core.StringPtr("searchString"),
				SourceIds:          []int64{int64(26)},
			}

			searchEmailRequestParamsBaseModel := &backuprecoveryv1.SearchEmailRequestParamsBase{
				AttendeesAddresses:        []string{"attendee1@domain.com"},
				BccRecipientAddresses:     []string{"bccrecipient@domain.com"},
				CcRecipientAddresses:      []string{"ccrecipient@domain.com"},
				CreatedEndTimeSecs:        core.Int64Ptr(int64(26)),
				CreatedStartTimeSecs:      core.Int64Ptr(int64(26)),
				DueDateEndTimeSecs:        core.Int64Ptr(int64(26)),
				DueDateStartTimeSecs:      core.Int64Ptr(int64(26)),
				EmailAddress:              core.StringPtr("email@domain.com"),
				EmailSubject:              core.StringPtr("Email Subject"),
				FirstName:                 core.StringPtr("First Name"),
				FolderNames:               []string{"folder1"},
				HasAttachment:             core.BoolPtr(true),
				LastModifiedEndTimeSecs:   core.Int64Ptr(int64(26)),
				LastModifiedStartTimeSecs: core.Int64Ptr(int64(26)),
				LastName:                  core.StringPtr("Last Name"),
				MiddleName:                core.StringPtr("Middle Name"),
				OrganizerAddress:          core.StringPtr("organizer@domain.com"),
				ReceivedEndTimeSecs:       core.Int64Ptr(int64(26)),
				ReceivedStartTimeSecs:     core.Int64Ptr(int64(26)),
				RecipientAddresses:        []string{"recipient@domain.com"},
				SenderAddress:             core.StringPtr("sender@domain.com"),
				SourceEnvironment:         core.StringPtr("kO365"),
				TaskStatusTypes:           []string{"NotStarted"},
				Types:                     []string{"Email"},
			}

			o365SearchRequestParamsModel := &backuprecoveryv1.O365SearchRequestParams{
				DomainIds: []int64{int64(26)},
				GroupIds:  []int64{int64(26)},
				SiteIds:   []int64{int64(26)},
				TeamsIds:  []int64{int64(26)},
				UserIds:   []int64{int64(26)},
			}

			searchDocumentLibraryRequestParamsModel := &backuprecoveryv1.SearchDocumentLibraryRequestParams{
				CategoryTypes:         []string{"Document"},
				CreationEndTimeSecs:   core.Int64Ptr(int64(26)),
				CreationStartTimeSecs: core.Int64Ptr(int64(26)),
				IncludeFiles:          core.BoolPtr(true),
				IncludeFolders:        core.BoolPtr(true),
				O365Params:            o365SearchRequestParamsModel,
				OwnerNames:            []string{"ownerName1"},
				SearchString:          core.StringPtr("searchString"),
				SizeBytesLowerLimit:   core.Int64Ptr(int64(26)),
				SizeBytesUpperLimit:   core.Int64Ptr(int64(26)),
			}

			searchMsGroupsRequestParamsModel := &backuprecoveryv1.SearchMsGroupsRequestParams{
				MailboxParams: searchEmailRequestParamsBaseModel,
				O365Params:    o365SearchRequestParamsModel,
				SiteParams:    searchDocumentLibraryRequestParamsModel,
			}

			o365TeamsChannelsSearchRequestParamsModel := &backuprecoveryv1.O365TeamsChannelsSearchRequestParams{
				ChannelEmail:           core.StringPtr("channel@domain.com"),
				ChannelID:              core.StringPtr("channelId"),
				ChannelName:            core.StringPtr("channelName"),
				IncludePrivateChannels: core.BoolPtr(true),
				IncludePublicChannels:  core.BoolPtr(true),
			}

			searchMsTeamsRequestParamsModel := &backuprecoveryv1.SearchMsTeamsRequestParams{
				CategoryTypes:         []string{"Document"},
				ChannelNames:          []string{"channelName1"},
				ChannelParams:         o365TeamsChannelsSearchRequestParamsModel,
				CreationEndTimeSecs:   core.Int64Ptr(int64(26)),
				CreationStartTimeSecs: core.Int64Ptr(int64(26)),
				O365Params:            o365SearchRequestParamsModel,
				OwnerNames:            []string{"ownerName1"},
				SearchString:          core.StringPtr("searchString"),
				SizeBytesLowerLimit:   core.Int64Ptr(int64(26)),
				SizeBytesUpperLimit:   core.Int64Ptr(int64(26)),
				Types:                 []string{"Channel"},
			}

			searchPublicFolderRequestParamsModel := &backuprecoveryv1.SearchPublicFolderRequestParams{
				SearchString:          core.StringPtr("searchString"),
				Types:                 []string{"Calendar"},
				HasAttachment:         core.BoolPtr(true),
				SenderAddress:         core.StringPtr("sender@domain.com"),
				RecipientAddresses:    []string{"recipient@domain.com"},
				CcRecipientAddresses:  []string{"ccrecipient@domain.com"},
				BccRecipientAddresses: []string{"bccrecipient@domain.com"},
				ReceivedStartTimeSecs: core.Int64Ptr(int64(26)),
				ReceivedEndTimeSecs:   core.Int64Ptr(int64(26)),
			}

			searchSfdcRecordsRequestParamsModel := &backuprecoveryv1.SearchSfdcRecordsRequestParams{
				MutationTypes: []string{"All"},
				ObjectName:    core.StringPtr("objectName"),
				QueryString:   core.StringPtr("queryString"),
				SnapshotID:    core.StringPtr("snapshotId"),
			}

			udaOnPremSearchParamsModel := &backuprecoveryv1.UdaOnPremSearchParams{
				SearchString: core.StringPtr("searchString"),
				SourceIds:    []int64{int64(26)},
			}

			searchIndexedObjectsOptions := &backuprecoveryv1.SearchIndexedObjectsOptions{
				XIBMTenantID:            core.StringPtr("tenantId"),
				ObjectType:              core.StringPtr("Emails"),
				ProtectionGroupIds:      []string{"protectionGroupId1"},
				StorageDomainIds:        []int64{int64(26)},
				TenantID:                core.StringPtr("tenantId"),
				IncludeTenants:          core.BoolPtr(false),
				Tags:                    []string{"123:456:ABC-123", "123:456:ABC-456"},
				SnapshotTags:            []string{"123:456:DEF-123", "123:456:DEF-456"},
				MustHaveTagIds:          []string{"123:456:ABC-123"},
				MightHaveTagIds:         []string{"123:456:ABC-456"},
				MustHaveSnapshotTagIds:  []string{"123:456:DEF-123"},
				MightHaveSnapshotTagIds: []string{"123:456:DEF-456"},
				PaginationCookie:        core.StringPtr("paginationCookie"),
				Count:                   core.Int64Ptr(int64(38)),
				UseCachedData:           core.BoolPtr(true),
				CassandraParams:         cassandraOnPremSearchParamsModel,
				CouchbaseParams:         couchBaseOnPremSearchParamsModel,
				EmailParams:             searchEmailRequestParamsModel,
				ExchangeParams:          searchExchangeObjectsRequestParamsModel,
				FileParams:              searchFileRequestParamsModel,
				HbaseParams:             hbaseOnPremSearchParamsModel,
				HdfsParams:              hdfsOnPremSearchParamsModel,
				HiveParams:              hiveOnPremSearchParamsModel,
				MongodbParams:           mongoDbOnPremSearchParamsModel,
				MsGroupsParams:          searchMsGroupsRequestParamsModel,
				MsTeamsParams:           searchMsTeamsRequestParamsModel,
				OneDriveParams:          searchDocumentLibraryRequestParamsModel,
				PublicFolderParams:      searchPublicFolderRequestParamsModel,
				SfdcParams:              searchSfdcRecordsRequestParamsModel,
				SharepointParams:        searchDocumentLibraryRequestParamsModel,
				UdaParams:               udaOnPremSearchParamsModel,
			}

			searchIndexedObjectsResponse, response, err := backupRecoveryService.SearchIndexedObjects(searchIndexedObjectsOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(searchIndexedObjectsResponse).ToNot(BeNil())
		})
	})

	Describe(`SearchObjects - List Objects`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`SearchObjects(searchObjectsOptions *SearchObjectsOptions)`, func() {
			searchObjectsOptions := &backuprecoveryv1.SearchObjectsOptions{
				XIBMTenantID:                   core.StringPtr("tenantId"),
				RequestInitiatorType:           core.StringPtr("UIUser"),
				SearchString:                   core.StringPtr("searchString"),
				Environments:                   []string{"kPhysical"},
				ProtectionTypes:                []string{"kAgent"},
				ProtectionGroupIds:             []string{"protectionGroupId1"},
				ObjectIds:                      []int64{int64(26)},
				OsTypes:                        []string{"kLinux"},
				SourceIds:                      []int64{int64(26)},
				SourceUUIDs:                    []string{"sourceUuid1"},
				IsProtected:                    core.BoolPtr(true),
				IsDeleted:                      core.BoolPtr(true),
				LastRunStatusList:              []string{"Accepted"},
				ClusterIdentifiers:             []string{"clusterIdentifier1"},
				IncludeDeletedObjects:          core.BoolPtr(true),
				PaginationCookie:               core.StringPtr("paginationCookie"),
				Count:                          core.Int64Ptr(int64(38)),
				MustHaveTagIds:                 []string{"123:456:ABC-123"},
				MightHaveTagIds:                []string{"123:456:ABC-456"},
				MustHaveSnapshotTagIds:         []string{"123:456:DEF-123"},
				MightHaveSnapshotTagIds:        []string{"123:456:DEF-456"},
				TagSearchName:                  core.StringPtr("tagName"),
				TagNames:                       []string{"tag1"},
				TagTypes:                       []string{"System"},
				TagCategories:                  []string{"Security"},
				TagSubCategories:               []string{"Classification"},
				IncludeHeliosTagInfoForObjects: core.BoolPtr(true),
				ExternalFilters:                []string{"filter1"},
			}

			objectsSearchResponseBody, response, err := backupRecoveryService.SearchObjects(searchObjectsOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(objectsSearchResponseBody).ToNot(BeNil())
		})
	})

	Describe(`SearchProtectedObjects - List Protected Objects`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`SearchProtectedObjects(searchProtectedObjectsOptions *SearchProtectedObjectsOptions)`, func() {
			searchProtectedObjectsOptions := &backuprecoveryv1.SearchProtectedObjectsOptions{
				XIBMTenantID:            core.StringPtr("tenantId"),
				RequestInitiatorType:    core.StringPtr("UIUser"),
				SearchString:            core.StringPtr("searchString"),
				Environments:            []string{"kPhysical"},
				SnapshotActions:         []string{"RecoverVMs"},
				ObjectActionKey:         core.StringPtr("kPhysical"),
				ProtectionGroupIds:      []string{"protectionGroupId1"},
				ObjectIds:               []int64{int64(26)},
				SubResultSize:           core.Int64Ptr(int64(38)),
				FilterSnapshotFromUsecs: core.Int64Ptr(int64(26)),
				FilterSnapshotToUsecs:   core.Int64Ptr(int64(26)),
				OsTypes:                 []string{"kLinux"},
				SourceIds:               []int64{int64(26)},
				RunInstanceIds:          []int64{int64(26)},
				CdpProtectedOnly:        core.BoolPtr(true),
				UseCachedData:           core.BoolPtr(true),
			}

			protectedObjectsSearchResponse, response, err := backupRecoveryService.SearchProtectedObjects(searchProtectedObjectsOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(protectedObjectsSearchResponse).ToNot(BeNil())
		})
	})

	Describe(`GetSourceRegistrations - Get the list of Protection Source registrations`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetSourceRegistrations(getSourceRegistrationsOptions *GetSourceRegistrationsOptions)`, func() {
			getSourceRegistrationsOptions := &backuprecoveryv1.GetSourceRegistrationsOptions{
				XIBMTenantID:                         core.StringPtr("tenantId"),
				Ids:                                  []int64{int64(38)},
				IncludeSourceCredentials:             core.BoolPtr(true),
				EncryptionKey:                        core.StringPtr("encryptionKey"),
				UseCachedData:                        core.BoolPtr(true),
				IncludeExternalMetadata:              core.BoolPtr(true),
				IgnoreTenantMigrationInProgressCheck: core.BoolPtr(true),
			}

			sourceRegistrations, response, err := backupRecoveryService.GetSourceRegistrations(getSourceRegistrationsOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(sourceRegistrations).ToNot(BeNil())
		})
	})

	Describe(`GetProtectionSourceRegistration - Get a Protection Source registration`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetProtectionSourceRegistration(getProtectionSourceRegistrationOptions *GetProtectionSourceRegistrationOptions)`, func() {
			getProtectionSourceRegistrationOptions := &backuprecoveryv1.GetProtectionSourceRegistrationOptions{
				ID:                   &protectionSourceIdLink,
				XIBMTenantID:         core.StringPtr("tenantId"),
				RequestInitiatorType: core.StringPtr("UIUser"),
			}

			sourceRegistrationResponseParams, response, err := backupRecoveryService.GetProtectionSourceRegistration(getProtectionSourceRegistrationOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(sourceRegistrationResponseParams).ToNot(BeNil())
		})
	})

	Describe(`UpdateProtectionSourceRegistration - Update Protection Source registration`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`UpdateProtectionSourceRegistration(updateProtectionSourceRegistrationOptions *UpdateProtectionSourceRegistrationOptions)`, func() {
			connectionConfigModel := &backuprecoveryv1.ConnectionConfig{
				ConnectionID:           core.Int64Ptr(int64(26)),
				EntityID:               core.Int64Ptr(int64(26)),
				ConnectorGroupID:       core.Int64Ptr(int64(26)),
				DataSourceConnectionID: core.StringPtr("DatasourceConnectionId"),
			}

			keyValuePairModel := &backuprecoveryv1.KeyValuePair{
				Key:   core.StringPtr("configKey"),
				Value: core.StringPtr("configValue"),
			}

			physicalSourceRegistrationParamsModel := &backuprecoveryv1.PhysicalSourceRegistrationParams{
				Endpoint:      core.StringPtr("xxx.xx.xx.xx"),
				ForceRegister: core.BoolPtr(true),
				HostType:      core.StringPtr("kLinux"),
				PhysicalType:  core.StringPtr("kGroup"),
				Applications:  []string{"kSQL"},
			}

			updateProtectionSourceRegistrationOptions := &backuprecoveryv1.UpdateProtectionSourceRegistrationOptions{
				ID:                         &protectionSourceIdLink,
				XIBMTenantID:               core.StringPtr("tenantId"),
				Environment:                core.StringPtr("kPhysical"),
				Name:                       core.StringPtr("update-protection-source"),
				IsInternalEncrypted:        core.BoolPtr(true),
				EncryptionKey:              core.StringPtr("encryptionKey"),
				ConnectionID:               core.Int64Ptr(int64(26)),
				Connections:                []backuprecoveryv1.ConnectionConfig{*connectionConfigModel},
				ConnectorGroupID:           core.Int64Ptr(int64(26)),
				AdvancedConfigs:            []backuprecoveryv1.KeyValuePair{*keyValuePairModel},
				DataSourceConnectionID:     core.StringPtr("DatasourceConnectionId"),
				LastModifiedTimestampUsecs: core.Int64Ptr(int64(26)),
				PhysicalParams:             physicalSourceRegistrationParamsModel,
			}

			sourceRegistrationResponseParams, response, err := backupRecoveryService.UpdateProtectionSourceRegistration(updateProtectionSourceRegistrationOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(sourceRegistrationResponseParams).ToNot(BeNil())
		})
	})

	Describe(`PatchProtectionSourceRegistration - Patches a Protection Source`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`PatchProtectionSourceRegistration(patchProtectionSourceRegistrationOptions *PatchProtectionSourceRegistrationOptions)`, func() {
			patchProtectionSourceRegistrationOptions := &backuprecoveryv1.PatchProtectionSourceRegistrationOptions{
				ID:           &protectionSourceIdLink,
				XIBMTenantID: core.StringPtr("tenantId"),
				Environment:  core.StringPtr("kPhysical"),
			}

			sourceRegistrationResponseParams, response, err := backupRecoveryService.PatchProtectionSourceRegistration(patchProtectionSourceRegistrationOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(sourceRegistrationResponseParams).ToNot(BeNil())
		})
	})

	Describe(`RefreshProtectionSourceByID - Refresh a Protection Source`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`RefreshProtectionSourceByID(refreshProtectionSourceByIdOptions *RefreshProtectionSourceByIdOptions)`, func() {
			refreshProtectionSourceByIdOptions := &backuprecoveryv1.RefreshProtectionSourceByIdOptions{
				ID:           &protectionSourceIdLink,
				XIBMTenantID: core.StringPtr("tenantId"),
			}

			response, err := backupRecoveryService.RefreshProtectionSourceByID(refreshProtectionSourceByIdOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(204))
		})
	})

	Describe(`GetUsers - Get Users`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`GetUsers(getUsersOptions *GetUsersOptions)`, func() {
			getUsersOptions := &backuprecoveryv1.GetUsersOptions{
				SessionName:       core.StringPtr("MTczNjc0NzY1OHxEWDhFQVFMX2dBQUJFQUVRQUFELUFZWF9nQUFKQm5OMGNtbHVad3dLQUFoMWMyVnlibUZ0WlFaemRISnBibWNNQndBRllXUnRhVzRHYzNSeWFXNW5EQWNBQlhKdmJHVnpCbk4wY21sdVp3d1FBQTVEVDBoRlUwbFVXVjlCUkUxSlRnWnpkSEpwYm1jTUN3QUpjMmxrY3kxb1lYTm9Cbk4wY21sdVp3d3RBQ3RTYVV4ZmFqQmZOVGxxZFZJeWVIVlZhREJ2UVZGNlUxcEhTVWc1TlZVdFlVWTBjV1JNUjNaTk9VUTBCbk4wY21sdVp3d01BQXBwYmkxamJIVnpkR1Z5QkdKdmIyd0NBZ0FCQm5OMGNtbHVad3dMQUFsaGRYUm9MWFI1Y0dVR2MzUnlhVzVuREFNQUFURUdjM1J5YVc1bkRCRUFEMlY0Y0dseVlYUnBiMjR0ZEdsdFpRWnpkSEpwYm1jTURBQUtNVGN6Tmpnek5EQTFPQVp6ZEhKcGJtY01DZ0FJZFhObGNpMXphV1FHYzNSeWFXNW5EQ0FBSGxNdE1TMHhNREF0TWpFdE16YzRNVFkyTXpVdE1qUXhPRFk1TXpVdE1RWnpkSEpwYm1jTUNBQUdaRzl0WVdsdUJuTjBjbWx1Wnd3SEFBVk1UME5CVEFaemRISnBibWNNQ0FBR2JHOWpZV3hsQm5OMGNtbHVad3dIQUFWbGJpMTFjdz09fGXFZlPU_3Nl46_gPKAw619qs6Pl7PX453Y_lf5BvBBo"),
				TenantIds:         []string{"testString"},
				AllUnderHierarchy: core.BoolPtr(true),
				Usernames:         []string{"testString"},
				EmailAddresses:    []string{"testString"},
				Domain:            core.StringPtr("testString"),
				PartialMatch:      core.BoolPtr(true),
			}

			userDetails, response, err := backupRecoveryService.GetUsers(getUsersOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(userDetails).ToNot(BeNil())
		})
	})

	Describe(`UpdateUser - Update User`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`UpdateUser(updateUserOptions *UpdateUserOptions)`, func() {
			adUserInfoModel := &backuprecoveryv1.AdUserInfo{
				GroupSids:      []string{"testString"},
				Groups:         []string{"testString"},
				IsFloatingUser: core.BoolPtr(true),
			}

			auditLogSettingsModel := &backuprecoveryv1.AuditLogSettings{
				ReadLogging: core.BoolPtr(true),
			}

			userClusterIdentifierModel := &backuprecoveryv1.UserClusterIdentifier{
				ClusterID:            core.Int64Ptr(int64(26)),
				ClusterIncarnationID: core.Int64Ptr(int64(26)),
			}

			googleAccountInfoModel := &backuprecoveryv1.GoogleAccountInfo{
				AccountID: core.StringPtr("testString"),
				UserID:    core.StringPtr("testString"),
			}

			idpUserInfoModel := &backuprecoveryv1.IdpUserInfo{
				GroupSids:      []string{"testString"},
				Groups:         []string{"testString"},
				IdpID:          core.Int64Ptr(int64(26)),
				IsFloatingUser: core.BoolPtr(true),
				IssuerID:       core.StringPtr("testString"),
				UserID:         core.StringPtr("testString"),
				Vendor:         core.StringPtr("testString"),
			}

			mfaInfoModel := &backuprecoveryv1.MfaInfo{
				IsUserExemptFromMfa: core.BoolPtr(true),
			}

			tenantConfigModel := &backuprecoveryv1.TenantConfig{
				BifrostEnabled:    core.BoolPtr(true),
				IsManagedOnHelios: core.BoolPtr(true),
				Name:              core.StringPtr("testString"),
				Restricted:        core.BoolPtr(true),
				Roles:             []string{"testString"},
				TenantID:          core.StringPtr("testString"),
			}

			usersPreferencesModel := &backuprecoveryv1.UsersPreferences{
				Locale: core.StringPtr("testString"),
			}

			userProfileModel := &backuprecoveryv1.UserProfile{
				ClusterIdentifiers: []backuprecoveryv1.UserClusterIdentifier{*userClusterIdentifierModel},
				IsActive:           core.BoolPtr(true),
				IsDeleted:          core.BoolPtr(true),
				RegionIds:          []string{"testString"},
				TenantID:           core.StringPtr("testString"),
				TenantName:         core.StringPtr("testString"),
				TenantType:         core.StringPtr("Dmaas"),
			}

			salesforceAccountInfoModel := &backuprecoveryv1.SalesforceAccountInfo{
				AccountID:               core.StringPtr("testString"),
				HeliosAccessGrantStatus: core.StringPtr("testString"),
				IsDGaaSUser:             core.BoolPtr(true),
				IsDMaaSUser:             core.BoolPtr(true),
				IsDRaaSUser:             core.BoolPtr(true),
				IsRPaaSUser:             core.BoolPtr(true),
				IsSalesUser:             core.BoolPtr(true),
				IsSupportUser:           core.BoolPtr(true),
				UserID:                  core.StringPtr("testString"),
			}

			spogContextModel := &backuprecoveryv1.SpogContext{
				PrimaryClusterID:       core.Int64Ptr(int64(26)),
				PrimaryClusterUserSid:  core.StringPtr("testString"),
				PrimaryClusterUsername: core.StringPtr("testString"),
			}

			classificationInfoModel := &backuprecoveryv1.ClassificationInfo{
				EndDate:     core.StringPtr("testString"),
				IsActive:    core.BoolPtr(true),
				IsFreeTrial: core.BoolPtr(true),
				StartDate:   core.StringPtr("testString"),
			}

			tieringInfoModel := &backuprecoveryv1.TieringInfo{
				BackendTiering:  core.BoolPtr(true),
				FrontendTiering: core.BoolPtr(true),
				MaxRetention:    core.Int64Ptr(int64(26)),
			}

			dataProtectInfoModel := &backuprecoveryv1.DataProtectInfo{
				EndDate:                core.StringPtr("testString"),
				IsActive:               core.BoolPtr(true),
				IsFreeTrial:            core.BoolPtr(true),
				IsAwsSubscription:      core.BoolPtr(true),
				IsCohesitySubscription: core.BoolPtr(true),
				Quantity:               core.Int64Ptr(int64(26)),
				StartDate:              core.StringPtr("testString"),
				Tiering:                tieringInfoModel,
			}

			dataProtectAzureInfoModel := &backuprecoveryv1.DataProtectAzureInfo{
				EndDate:     core.StringPtr("testString"),
				IsActive:    core.BoolPtr(true),
				IsFreeTrial: core.BoolPtr(true),
				Quantity:    core.Int64Ptr(int64(26)),
				StartDate:   core.StringPtr("testString"),
				Tiering:     tieringInfoModel,
			}

			fortKnoxInfoModel := &backuprecoveryv1.FortKnoxInfo{
				EndDate:     core.StringPtr("testString"),
				IsActive:    core.BoolPtr(true),
				IsFreeTrial: core.BoolPtr(true),
				Quantity:    core.Int64Ptr(int64(26)),
				StartDate:   core.StringPtr("testString"),
			}

			subscriptionInfoModel := &backuprecoveryv1.SubscriptionInfo{
				Classification:    classificationInfoModel,
				DataProtect:       dataProtectInfoModel,
				DataProtectAzure:  dataProtectAzureInfoModel,
				FortKnoxAzureCool: fortKnoxInfoModel,
				FortKnoxAzureHot:  fortKnoxInfoModel,
				FortKnoxCold:      fortKnoxInfoModel,
				Ransomware:        fortKnoxInfoModel,
				SiteContinuity:    classificationInfoModel,
				ThreatProtection:  classificationInfoModel,
			}

			tenantAccessesModel := &backuprecoveryv1.TenantAccesses{
				ClusterIdentifiers:   []backuprecoveryv1.UserClusterIdentifier{*userClusterIdentifierModel},
				CreatedTimeMsecs:     core.Int64Ptr(int64(26)),
				EffectiveTimeMsecs:   core.Int64Ptr(int64(26)),
				ExpiredTimeMsecs:     core.Int64Ptr(int64(26)),
				IsAccessActive:       core.BoolPtr(true),
				IsActive:             core.BoolPtr(true),
				IsDeleted:            core.BoolPtr(true),
				LastUpdatedTimeMsecs: core.Int64Ptr(int64(26)),
				Roles:                []string{"testString"},
				TenantID:             core.StringPtr("testString"),
				TenantName:           core.StringPtr("testString"),
				TenantType:           core.StringPtr("Dmaas"),
			}

			updateUserOptions := &backuprecoveryv1.UpdateUserOptions{
				SessionName:                  core.StringPtr("MTczNjc0NzY1OHxEWDhFQVFMX2dBQUJFQUVRQUFELUFZWF9nQUFKQm5OMGNtbHVad3dLQUFoMWMyVnlibUZ0WlFaemRISnBibWNNQndBRllXUnRhVzRHYzNSeWFXNW5EQWNBQlhKdmJHVnpCbk4wY21sdVp3d1FBQTVEVDBoRlUwbFVXVjlCUkUxSlRnWnpkSEpwYm1jTUN3QUpjMmxrY3kxb1lYTm9Cbk4wY21sdVp3d3RBQ3RTYVV4ZmFqQmZOVGxxZFZJeWVIVlZhREJ2UVZGNlUxcEhTVWc1TlZVdFlVWTBjV1JNUjNaTk9VUTBCbk4wY21sdVp3d01BQXBwYmkxamJIVnpkR1Z5QkdKdmIyd0NBZ0FCQm5OMGNtbHVad3dMQUFsaGRYUm9MWFI1Y0dVR2MzUnlhVzVuREFNQUFURUdjM1J5YVc1bkRCRUFEMlY0Y0dseVlYUnBiMjR0ZEdsdFpRWnpkSEpwYm1jTURBQUtNVGN6Tmpnek5EQTFPQVp6ZEhKcGJtY01DZ0FJZFhObGNpMXphV1FHYzNSeWFXNW5EQ0FBSGxNdE1TMHhNREF0TWpFdE16YzRNVFkyTXpVdE1qUXhPRFk1TXpVdE1RWnpkSEpwYm1jTUNBQUdaRzl0WVdsdUJuTjBjbWx1Wnd3SEFBVk1UME5CVEFaemRISnBibWNNQ0FBR2JHOWpZV3hsQm5OMGNtbHVad3dIQUFWbGJpMTFjdz09fGXFZlPU_3Nl46_gPKAw619qs6Pl7PX453Y_lf5BvBBo"),
				AdUserInfo:                   adUserInfoModel,
				AdditionalGroupNames:         []string{"testString"},
				AllowDsoModify:               core.BoolPtr(true),
				AuditLogSettings:             auditLogSettingsModel,
				AuthenticationType:           core.StringPtr("kAuthLocal"),
				ClusterIdentifiers:           []backuprecoveryv1.UserClusterIdentifier{*userClusterIdentifierModel},
				CreatedTimeMsecs:             core.Int64Ptr(int64(26)),
				CurrentPassword:              core.StringPtr("testString"),
				Description:                  core.StringPtr("testString"),
				Domain:                       core.StringPtr("testString"),
				EffectiveTimeMsecs:           core.Int64Ptr(int64(26)),
				EmailAddress:                 core.StringPtr("testString"),
				ExpiredTimeMsecs:             core.Int64Ptr(int64(26)),
				ForcePasswordChange:          core.BoolPtr(true),
				GoogleAccount:                googleAccountInfoModel,
				IdpUserInfo:                  idpUserInfoModel,
				IntercomMessengerToken:       core.StringPtr("testString"),
				IsAccountLocked:              core.BoolPtr(true),
				IsActive:                     core.BoolPtr(true),
				LastSuccessfulLoginTimeMsecs: core.Int64Ptr(int64(26)),
				LastUpdatedTimeMsecs:         core.Int64Ptr(int64(26)),
				MfaInfo:                      mfaInfoModel,
				MfaMethods:                   []string{"testString"},
				ObjectClass:                  core.StringPtr("testString"),
				OrgMembership:                []backuprecoveryv1.TenantConfig{*tenantConfigModel},
				Password:                     core.StringPtr("testString"),
				Preferences:                  usersPreferencesModel,
				PreviousLoginTimeMsecs:       core.Int64Ptr(int64(26)),
				PrimaryGroupName:             core.StringPtr("testString"),
				PrivilegeIds:                 []string{"kPrincipalView"},
				Profiles:                     []backuprecoveryv1.UserProfile{*userProfileModel},
				Restricted:                   core.BoolPtr(true),
				Roles:                        []string{"testString"},
				S3AccessKeyID:                core.StringPtr("testString"),
				S3AccountID:                  core.StringPtr("testString"),
				S3SecretKey:                  core.StringPtr("testString"),
				SalesforceAccount:            salesforceAccountInfoModel,
				Sid:                          core.StringPtr("testString"),
				SpogContext:                  spogContextModel,
				SubscriptionInfo:             subscriptionInfoModel,
				TenantAccesses:               []backuprecoveryv1.TenantAccesses{*tenantAccessesModel},
				TenantID:                     core.StringPtr("testString"),
				Username:                     core.StringPtr("testString"),
			}

			userDetails, response, err := backupRecoveryService.UpdateUser(updateUserOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(200))
			Expect(userDetails).ToNot(BeNil())
		})
	})

	Describe(`DeleteDataSourceConnection - Delete the data-source connection specified by the ID in the request path`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`DeleteDataSourceConnection(deleteDataSourceConnectionOptions *DeleteDataSourceConnectionOptions)`, func() {
			deleteDataSourceConnectionOptions := &backuprecoveryv1.DeleteDataSourceConnectionOptions{
				ConnectionID: &connectionIdLink,
				XIBMTenantID: core.StringPtr("tenantId"),
			}

			response, err := backupRecoveryService.DeleteDataSourceConnection(deleteDataSourceConnectionOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(204))
		})
	})

	Describe(`DeleteProtectionPolicy - Delete a Protection Policy`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`DeleteProtectionPolicy(deleteProtectionPolicyOptions *DeleteProtectionPolicyOptions)`, func() {
			deleteProtectionPolicyOptions := &backuprecoveryv1.DeleteProtectionPolicyOptions{
				ID:           &policyIdLink,
				XIBMTenantID: core.StringPtr("tenantId"),
			}

			response, err := backupRecoveryService.DeleteProtectionPolicy(deleteProtectionPolicyOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(204))
		})
	})

	Describe(`DeleteProtectionGroup - Delete a Protection Group`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`DeleteProtectionGroup(deleteProtectionGroupOptions *DeleteProtectionGroupOptions)`, func() {
			deleteProtectionGroupOptions := &backuprecoveryv1.DeleteProtectionGroupOptions{
				ID:              &protectionGroupIdLink,
				XIBMTenantID:    core.StringPtr("tenantId"),
				DeleteSnapshots: core.BoolPtr(true),
			}

			response, err := backupRecoveryService.DeleteProtectionGroup(deleteProtectionGroupOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(204))
		})
	})

	Describe(`DeleteProtectionSourceRegistration - Delete Protection Source Registration`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`DeleteProtectionSourceRegistration(deleteProtectionSourceRegistrationOptions *DeleteProtectionSourceRegistrationOptions)`, func() {
			deleteProtectionSourceRegistrationOptions := &backuprecoveryv1.DeleteProtectionSourceRegistrationOptions{
				ID:           &protectionSourceIdLink,
				XIBMTenantID: core.StringPtr("tenantId"),
			}

			response, err := backupRecoveryService.DeleteProtectionSourceRegistration(deleteProtectionSourceRegistrationOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(204))
		})
	})

	Describe(`DeleteDataSourceConnector - Delete a data-source connector using its ID`, func() {
		BeforeEach(func() {
			shouldSkipTest()
		})
		It(`DeleteDataSourceConnector(deleteDataSourceConnectorOptions *DeleteDataSourceConnectorOptions)`, func() {
			deleteDataSourceConnectorOptions := &backuprecoveryv1.DeleteDataSourceConnectorOptions{
				ConnectorID:  core.StringPtr("connectorId"),
				XIBMTenantID: core.StringPtr("tenantId"),
			}

			response, err := backupRecoveryService.DeleteDataSourceConnector(deleteDataSourceConnectorOptions)
			Expect(err).To(BeNil())
			Expect(response.StatusCode).To(Equal(204))
		})
	})
})

//
// Utility functions are declared in the unit test file
//
