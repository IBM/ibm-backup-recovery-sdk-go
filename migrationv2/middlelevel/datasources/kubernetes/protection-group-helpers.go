/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package kubernetes

import (
	"fmt"
	"strings"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/ibm-backup-recovery-sdk-go/backuprecoveryv1"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/types"
)

func parseNamespaceList(ns string) []string {
	if ns == "" {
		return []string{}
	}

	parts := strings.Split(ns, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return parts
}

// Returns both IDs and names
func (k *KubernetesDataSource) fetchNamespaceIDsWithNames(registrationID int64, includeList, excludeList []string) ([]types.NamespaceData, error) {
	tenantId := k.brsClient.GetTenantId()
	listOptions := &backuprecoveryv1.ListProtectionSourcesOptions{
		XIBMTenantID: core.StringPtr(tenantId),
		ID:           core.Int64Ptr(registrationID),
	}

	sources, _, err := k.brsClient.GetBRSClient().ListProtectionSources(listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list protection sources: %w", err)
	}

	// Check len() before slice index access
	if sources == nil || len(sources) == 0 {
		return nil, fmt.Errorf("no protection sources found for registration ID: %d", registrationID)
	}

	var namespaceList []types.NamespaceData

	// Nil check before accessing Nodes
	if sources[0].Nodes == nil {
		return namespaceList, nil
	}

	for _, topNode := range sources[0].Nodes {
		// Nil check before accessing child Nodes
		if topNode.Nodes == nil {
			continue
		}

		for _, child := range topNode.Nodes {
			// Nil checks before pointer dereferences
			if child.ProtectionSource == nil {
				continue
			}

			src := child.ProtectionSource.KubernetesProtectionSource
			if src == nil || src.Type == nil || *src.Type != backuprecoveryv1.KubernetesProtectionSource_Type_Knamespace {
				continue
			}

			// Nil check before dereferencing Name
			if src.Name == nil {
				continue
			}
			nsName := *src.Name
			if (len(includeList) > 0 && !isIncluded(nsName, includeList)) || isExcluded(nsName, excludeList) {
				continue
			}

			// Nil check before dereferencing ID
			if child.ProtectionSource.ID == nil {
				continue
			}

			ns := types.NamespaceData{
				Id:     *child.ProtectionSource.ID,
				Name:   nsName,
				PvcIDs: make(map[string]int64),
			}

			// Discover PVC IDs under namespace
			if child.Nodes != nil {
				for _, pvcNode := range child.Nodes {
					// Nil checks before pointer dereferences
					if pvcNode.ProtectionSource == nil {
						continue
					}
					pvcSrc := pvcNode.ProtectionSource.KubernetesProtectionSource
					if pvcSrc != nil && pvcSrc.Type != nil && *pvcSrc.Type == backuprecoveryv1.KubernetesSourceRegistrationParams_KubernetesType_Kpersistentvolumeclaim {
						if pvcSrc.Name != nil && pvcNode.ProtectionSource.ID != nil {
							ns.PvcIDs[*pvcSrc.Name] = *pvcNode.ProtectionSource.ID
						}
					}
				}
			}

			namespaceList = append(namespaceList, ns)
		}
	}

	return namespaceList, nil
}

func isExcluded(name string, excludeList []string) bool {
	for _, ex := range excludeList {
		if name == ex {
			return true
		}
	}

	return false
}

func isIncluded(name string, includeList []string) bool {
	for _, inc := range includeList {
		if name == inc {
			return true
		}
	}

	return false
}

func (k *KubernetesDataSource) GetK8Object(uniqueNameSpaces []types.NamespaceData) *backuprecoveryv1.KubernetesProtectionGroupParams {
	if k == nil || k.KubernetesProtectionParams == nil {
		return nil
	}

	var (
		objs      []backuprecoveryv1.KubernetesProtectionGroupObjectParams
		k8sParams *backuprecoveryv1.KubernetesProtectionGroupParams
	)

	for _, ns := range uniqueNameSpaces {
		var quiesceGroups []backuprecoveryv1.QuiesceGroup
		var failBackupOnHookFailure *bool
		var workloadIncludeParams *backuprecoveryv1.KubernetesFilterParams
		var resourceInclusion []string
		var resourceExclusion []string
		var pvcInclusion []backuprecoveryv1.KubernetesPvcInfo
		var pvcExclusion []backuprecoveryv1.KubernetesPvcInfo

		for _, nsSetting := range k.KubernetesProtectionParams.NamespacesSetting {
			if nsSetting.Namespace != ns.Name {
				continue
			}
			failBackupOnHookFailure = core.BoolPtr(
				nsSetting.Hooks.FailBackupIfHookFailed,
			)
			if strings.TrimSpace(k.KubernetesProtectionParams.IncludeAppLabels) != "" {
				workloadIncludeParams = parseLabels(
					k.KubernetesProtectionParams.IncludeAppLabels,
					"AND",
				)
			}
			if nsSetting.Resources.Inclusion != nil {
				resourceInclusion = nsSetting.Resources.Inclusion.ResourceTypes
			}
			if nsSetting.Resources.Exclusion != nil {
				resourceExclusion = nsSetting.Resources.Exclusion.ResourceTypes
			}

			pvcInclusion = MapPvcInclusion(nsSetting, ns)
			pvcExclusion = MapPvcExclusion(nsSetting, ns)

			if nsSetting.Hooks.RulesApplyMode != "" {
				quiesceGroups = MapRules(nsSetting, ns)
			}

		}

		objs = append(objs, backuprecoveryv1.KubernetesProtectionGroupObjectParams{
			ID:                      core.Int64Ptr(ns.Id),
			IncludedResources:       resourceInclusion,
			ExcludedResources:       resourceExclusion,
			IncludePvcs:             pvcInclusion,
			ExcludePvcs:             pvcExclusion,
			FailBackupOnHookFailure: failBackupOnHookFailure,
			IncludeParams:           workloadIncludeParams,
			QuiesceGroups:           quiesceGroups,
		})
	}

	k8sParams = &backuprecoveryv1.KubernetesProtectionGroupParams{
		Objects: objs,
	}

	// Only set Settings-related fields if Settings is not nil
	if k.KubernetesProtectionParams.Settings != nil {
		k8sParams.LeverageCSISnapshot = core.BoolPtr(k.KubernetesProtectionParams.Settings.CSISnapshot)
		k8sParams.IncludeParams = parseLabels(k.KubernetesProtectionParams.Settings.Labels.PersistentVolumeClaim.Inclusion,
			k.KubernetesProtectionParams.Settings.Labels.PersistentVolumeClaim.LogicRule)
		k8sParams.ExcludeParams = parseLabels(k.KubernetesProtectionParams.Settings.Labels.PersistentVolumeClaim.Exclusion,
			k.KubernetesProtectionParams.Settings.Labels.PersistentVolumeClaim.LogicRule)
	}

	return k8sParams
}

func MapPvcInclusion(nsSetting types.NamespacesSetting, ns types.NamespaceData) []backuprecoveryv1.KubernetesPvcInfo {

	var pvcInclusion []backuprecoveryv1.KubernetesPvcInfo
	for _, pvcName := range nsSetting.PersistentVolumeClaims.Inclusion {
		pvcName = strings.TrimSpace(pvcName)
		if pvcName == "" {
			continue
		}

		if pvcID, ok := ns.PvcIDs[pvcName]; ok {
			// NEW BEHAVIOR — use ID if discovered
			pvcInclusion = append(pvcInclusion, backuprecoveryv1.KubernetesPvcInfo{
				ID: core.Int64Ptr(pvcID),
			})
		} else {
			// FALLBACK — old behavior
			pvcInclusion = append(pvcInclusion, backuprecoveryv1.KubernetesPvcInfo{
				Name: core.StringPtr(pvcName),
			})
		}
	}
	return pvcInclusion
}

func MapPvcExclusion(nsSetting types.NamespacesSetting, ns types.NamespaceData) []backuprecoveryv1.KubernetesPvcInfo {

	var pvcExclusion []backuprecoveryv1.KubernetesPvcInfo
	for _, pvcName := range nsSetting.PersistentVolumeClaims.Exclusion {
		pvcName = strings.TrimSpace(pvcName)
		if pvcName == "" {
			continue
		}

		if pvcID, ok := ns.PvcIDs[pvcName]; ok {
			// NEW BEHAVIOR — use ID if discovered
			pvcExclusion = append(pvcExclusion, backuprecoveryv1.KubernetesPvcInfo{
				ID: core.Int64Ptr(pvcID),
			})
		} else {
			// FALLBACK — old behavior
			pvcExclusion = append(pvcExclusion, backuprecoveryv1.KubernetesPvcInfo{
				Name: core.StringPtr(pvcName),
			})
		}
	}
	return pvcExclusion
}

func MapRules(nsSetting types.NamespacesSetting, ns types.NamespaceData) []backuprecoveryv1.QuiesceGroup {
	var quiesceGroups []backuprecoveryv1.QuiesceGroup
	quiesceRules := make([]backuprecoveryv1.QuiesceRule, 0, len(nsSetting.Hooks.Rules))
	for _, rule := range nsSetting.Hooks.Rules {
		podLabels := parsePodLabels(rule.Rule.PodLabels)
		var preHooks []backuprecoveryv1.KubernetesHook
		var postHooks []backuprecoveryv1.KubernetesHook
		// PRE-SCRIPT (if present)
		if strings.TrimSpace(rule.Rule.PreScript) != "" {
			preHook := backuprecoveryv1.KubernetesHook{
				Commands: []string{rule.Rule.PreScript},
			}
			// Add container if provided
			if strings.TrimSpace(rule.Rule.Container) != "" {
				preHook.Container = core.StringPtr(rule.Rule.Container)
			}
			preHooks = append(preHooks, preHook)
		}
		// POST-SCRIPT (if present)
		if strings.TrimSpace(rule.Rule.PostScript) != "" {
			postHook := backuprecoveryv1.KubernetesHook{
				Commands: []string{rule.Rule.PostScript},
			}
			// Add container if provided
			if strings.TrimSpace(rule.Rule.Container) != "" {
				postHook.Container = core.StringPtr(rule.Rule.Container)
			}
			postHooks = append(postHooks, postHook)
		}
		quiesceRules = append(quiesceRules, backuprecoveryv1.QuiesceRule{
			PodSelectorLabels: podLabels,
			PreSnapshotHooks:  preHooks,
			PostSnapshotHooks: postHooks,
		})
	}
	quiesceGroups = append(quiesceGroups, backuprecoveryv1.QuiesceGroup{
		QuiesceMode: core.StringPtr(nsSetting.Hooks.RulesApplyMode),
		//FailBackupIfHookFailed: core.BoolPtr(nsSetting.Hooks.FailBackupIfHookFailed),
		QuiesceRules: quiesceRules,
	})
	return quiesceGroups
}

func parseLabels(input string, logicRule string) *backuprecoveryv1.KubernetesFilterParams {
	var labels []backuprecoveryv1.KubernetesLabel
	labelCombinationMethod := "OR"
	if logicRule == "MatchAll" {
		labelCombinationMethod = "AND"
	}
	if strings.TrimSpace(input) == "" {
		return &backuprecoveryv1.KubernetesFilterParams{
			LabelCombinationMethod: core.StringPtr(labelCombinationMethod),
			LabelVector:            labels,
		}
	}

	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		var key, value string

		// Support both ":" and "=" separators
		if strings.Contains(pair, ":") {
			parts := strings.SplitN(pair, ":", 2)
			key = strings.TrimSpace(parts[0])
			value = strings.TrimSpace(parts[1])
		} else if strings.Contains(pair, "=") {
			parts := strings.SplitN(pair, "=", 2)
			key = strings.TrimSpace(parts[0])
			value = strings.TrimSpace(parts[1])
		} else {
			// Skip invalid entries silently or log if needed
			continue
		}

		if key != "" {
			labels = append(labels, backuprecoveryv1.KubernetesLabel{
				Key:   core.StringPtr(key),
				Value: core.StringPtr(value),
			})
		}
	}
	return &backuprecoveryv1.KubernetesFilterParams{
		LabelCombinationMethod: core.StringPtr(labelCombinationMethod),
		LabelVector:            labels,
	}
}

func parsePodLabels(labelStr string) []backuprecoveryv1.KubernetesLabel {
	labelStr = strings.TrimSpace(labelStr)
	if labelStr == "" {
		return nil
	}

	var labels []backuprecoveryv1.KubernetesLabel

	for _, pair := range strings.Split(labelStr, ",") {
		pair = strings.TrimSpace(pair)

		var key, value string

		if strings.Contains(pair, "=") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key = strings.TrimSpace(parts[0])
			value = strings.TrimSpace(parts[1])
		} else if strings.Contains(pair, ":") {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key = strings.TrimSpace(parts[0])
			value = strings.TrimSpace(parts[1])
		} else {
			continue
		}

		if key == "" || value == "" {
			continue
		}

		labels = append(labels, backuprecoveryv1.KubernetesLabel{
			Key:   core.StringPtr(key),
			Value: core.StringPtr(value),
		})
	}

	if len(labels) == 0 {
		return nil
	}

	return labels
}

func uniqueByID(data []types.NamespaceData) []types.NamespaceData {
	seen := make(map[int64]bool)
	var result []types.NamespaceData

	for _, item := range data {
		if !seen[item.Id] {
			seen[item.Id] = true
			result = append(result, item)
		}
	}

	return result
}
