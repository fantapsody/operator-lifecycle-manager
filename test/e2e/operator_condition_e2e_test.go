package e2e

import (
	"context"

	"github.com/blang/semver/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	operatorsv2 "github.com/operator-framework/api/pkg/operators/v2"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry"
)

var _ = Describe("Operator Condition", func() {
	AfterEach(func() {
		TearDown(testNamespace)
	})

	It("OperatorCondition Upgradeable type and overrides", func() {
		By("This test proves that an operator can upgrade successfully when" +
			" Upgrade condition type is set in OperatorCondition spec. Plus, an operator" +
			" chooses not to use OperatorCondition, the upgrade process will proceed as" +
			" expected. The overrides spec in OperatorCondition can be used to override" +
			" the conditions spec. The overrides spec will remain in place until" +
			" they are unset.")
		c := newKubeClient()
		crc := newCRClient()

		// Create a catalog for csvA, csvB, and csvD
		pkgA := genName("a-")
		pkgB := genName("b-")
		pkgD := genName("d-")
		pkgAStable := pkgA + "-stable"
		pkgBStable := pkgB + "-stable"
		pkgDStable := pkgD + "-stable"
		stableChannel := "stable"
		strategyA := newNginxInstallStrategy(pkgAStable, nil, nil)
		strategyB := newNginxInstallStrategy(pkgBStable, nil, nil)
		strategyD := newNginxInstallStrategy(pkgDStable, nil, nil)
		crd := newCRD(genName(pkgA))
		csvA := newCSV(pkgAStable, testNamespace, "", semver.MustParse("0.1.0"), []apiextensions.CustomResourceDefinition{crd}, nil, &strategyA)
		csvB := newCSV(pkgBStable, testNamespace, pkgAStable, semver.MustParse("0.2.0"), []apiextensions.CustomResourceDefinition{crd}, nil, &strategyB)
		csvD := newCSV(pkgDStable, testNamespace, pkgBStable, semver.MustParse("0.3.0"), []apiextensions.CustomResourceDefinition{crd}, nil, &strategyD)

		// Create the initial catalogsources
		manifests := []registry.PackageManifest{
			{
				PackageName: pkgA,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: pkgAStable},
				},
				DefaultChannelName: stableChannel,
			},
		}

		catalog := genName("catalog-")
		_, cleanupCatalogSource := createInternalCatalogSource(c, crc, catalog, testNamespace, manifests, []apiextensions.CustomResourceDefinition{crd}, []operatorsv1alpha1.ClusterServiceVersion{csvA})
		defer cleanupCatalogSource()
		_, err := fetchCatalogSourceOnStatus(crc, catalog, testNamespace, catalogSourceRegistryPodSynced)
		subName := genName("sub-")
		cleanupSub := createSubscriptionForCatalog(crc, testNamespace, subName, catalog, pkgA, stableChannel, pkgAStable, operatorsv1alpha1.ApprovalAutomatic)
		defer cleanupSub()

		// Await csvA's success
		_, err = awaitCSV(crc, testNamespace, csvA.GetName(), csvSucceededChecker)
		require.NoError(GinkgoT(), err)

		// Get the OperatorCondition for csvA and report that it is not upgradeable
		var cond *operatorsv2.OperatorCondition
		upgradeableFalseCondition := metav1.Condition{
			Type:               operatorsv2.Upgradeable,
			Status:             metav1.ConditionFalse,
			Reason:             "test",
			Message:            "test",
			LastTransitionTime: metav1.Now(),
		}

		var currentGen int64
		Eventually(func() error {
			cond, err := crc.OperatorsV2().OperatorConditions(testNamespace).Get(context.TODO(), csvA.GetName(), metav1.GetOptions{})
			if err != nil {
				return err
			}
			currentGen = cond.ObjectMeta.GetGeneration()
			upgradeableFalseCondition.ObservedGeneration = currentGen
			meta.SetStatusCondition(&cond.Spec.Conditions, upgradeableFalseCondition)
			_, err = crc.OperatorsV2().OperatorConditions(testNamespace).Update(context.TODO(), cond, metav1.UpdateOptions{})
			return err
		}, pollInterval, pollDuration).Should(Succeed())

		// Update the catalogsources
		manifests = []registry.PackageManifest{
			{
				PackageName: pkgA,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: pkgBStable},
				},
				DefaultChannelName: stableChannel,
			},
		}
		updateInternalCatalog(GinkgoT(), c, crc, catalog, testNamespace, []apiextensions.CustomResourceDefinition{crd}, []operatorsv1alpha1.ClusterServiceVersion{csvA, csvB}, manifests)

		// Attempt to get the catalog source before creating install plan(s)
		_, err = fetchCatalogSourceOnStatus(crc, catalog, testNamespace, catalogSourceRegistryPodSynced)
		require.NoError(GinkgoT(), err)

		// csvB will be in Pending phase due to csvA reports Upgradeable=False condition
		fetchedCSV, err := fetchCSV(crc, csvB.GetName(), testNamespace, buildCSVReasonChecker(operatorsv1alpha1.CSVReasonOperatorConditionNotUpgradeable))
		require.NoError(GinkgoT(), err)
		require.Equal(GinkgoT(), fetchedCSV.Status.Phase, operatorsv1alpha1.CSVPhasePending)

		// Get the OperatorCondition for csvA and report that it is upgradeable, unblocking csvB
		upgradeableTrueCondition := metav1.Condition{
			Type:               operatorsv2.Upgradeable,
			Status:             metav1.ConditionTrue,
			Reason:             "test",
			Message:            "test",
			LastTransitionTime: metav1.Now(),
		}
		Eventually(func() error {
			cond, err = crc.OperatorsV2().OperatorConditions(testNamespace).Get(context.TODO(), csvA.GetName(), metav1.GetOptions{})
			if err != nil || currentGen == cond.ObjectMeta.GetGeneration() {
				return err
			}
			currentGen = cond.ObjectMeta.GetGeneration()
			upgradeableTrueCondition.ObservedGeneration = cond.ObjectMeta.GetGeneration()
			meta.SetStatusCondition(&cond.Spec.Conditions, upgradeableTrueCondition)
			_, err = crc.OperatorsV2().OperatorConditions(testNamespace).Update(context.TODO(), cond, metav1.UpdateOptions{})
			return err
		}, pollInterval, pollDuration).Should(Succeed())

		// Await csvB's success
		_, err = awaitCSV(crc, testNamespace, csvB.GetName(), csvSucceededChecker)
		require.NoError(GinkgoT(), err)

		// Get the OperatorCondition for csvB and purposedly change ObservedGeneration
		// to cause mismatch generation situation
		Eventually(func() error {
			cond, err = crc.OperatorsV2().OperatorConditions(testNamespace).Get(context.TODO(), csvB.GetName(), metav1.GetOptions{})
			if err != nil || currentGen == cond.ObjectMeta.GetGeneration() {
				return err
			}
			currentGen = cond.ObjectMeta.GetGeneration()
			upgradeableTrueCondition.ObservedGeneration = currentGen + 1
			meta.SetStatusCondition(&cond.Status.Conditions, upgradeableTrueCondition)
			_, err = crc.OperatorsV2().OperatorConditions(testNamespace).UpdateStatus(context.TODO(), cond, metav1.UpdateOptions{})
			return err
		}, pollInterval, pollDuration).Should(Succeed())

		// Update the catalogsources
		manifests = []registry.PackageManifest{
			{
				PackageName: pkgA,
				Channels: []registry.PackageChannel{
					{Name: stableChannel, CurrentCSVName: pkgDStable},
				},
				DefaultChannelName: stableChannel,
			},
		}

		updateInternalCatalog(GinkgoT(), c, crc, catalog, testNamespace, []apiextensions.CustomResourceDefinition{crd}, []operatorsv1alpha1.ClusterServiceVersion{csvA, csvB, csvD}, manifests)
		// Attempt to get the catalog source before creating install plan(s)
		_, err = fetchCatalogSourceOnStatus(crc, catalog, testNamespace, catalogSourceRegistryPodSynced)
		require.NoError(GinkgoT(), err)

		// CSVD will be in Pending status due to overrides in csvB's condition
		fetchedCSV, err = fetchCSV(crc, csvD.GetName(), testNamespace, buildCSVReasonChecker(operatorsv1alpha1.CSVReasonOperatorConditionNotUpgradeable))
		require.NoError(GinkgoT(), err)
		require.Equal(GinkgoT(), fetchedCSV.Status.Phase, operatorsv1alpha1.CSVPhasePending)

		// Get the OperatorCondition for csvB and override the upgradeable false condition
		Eventually(func() error {
			cond, err = crc.OperatorsV2().OperatorConditions(testNamespace).Get(context.TODO(), csvB.GetName(), metav1.GetOptions{})
			if err != nil {
				return err
			}
			meta.SetStatusCondition(&cond.Spec.Overrides, upgradeableTrueCondition)
			// Update the condition
			_, err = crc.OperatorsV2().OperatorConditions(testNamespace).Update(context.TODO(), cond, metav1.UpdateOptions{})
			return err
		}, pollInterval, pollDuration).Should(Succeed())
		require.NoError(GinkgoT(), err)

		require.NoError(GinkgoT(), err)
		_, err = awaitCSV(crc, testNamespace, csvD.GetName(), csvSucceededChecker)
		require.NoError(GinkgoT(), err)
	})
})
