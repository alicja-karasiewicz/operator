package reconcile

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	vmv1beta1 "github.com/VictoriaMetrics/operator/api/operator/v1beta1"
	"github.com/VictoriaMetrics/operator/internal/controller/operator/factory/finalize"
	"github.com/VictoriaMetrics/operator/internal/controller/operator/factory/logger"
)

// VMServiceScrapeForCRD creates or updates given object
func VMServiceScrapeForCRD(ctx context.Context, rclient client.Client, vss *vmv1beta1.VMServiceScrape) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var existVSS vmv1beta1.VMServiceScrape
		err := rclient.Get(ctx, types.NamespacedName{Namespace: vss.Namespace, Name: vss.Name}, &existVSS)
		if err != nil {
			if errors.IsNotFound(err) {
				return rclient.Create(ctx, vss)
			}
			return err
		}
		if err := finalize.FreeIfNeeded(ctx, rclient, &existVSS); err != nil {
			return err
		}

		if equality.Semantic.DeepEqual(vss.Spec, existVSS.Spec) &&
			equality.Semantic.DeepEqual(vss.Labels, existVSS.Labels) &&
			equality.Semantic.DeepEqual(vss.Annotations, existVSS.Annotations) {
			return nil
		}
		// do not allow to add 3rd party annotations and labels
		// since operator owns this resource
		existVSS.Annotations = vss.Annotations
		existVSS.Spec = vss.Spec
		existVSS.Labels = vss.Labels
		logger.WithContext(ctx).Info("updating vmservicescrape for CRD object")

		return rclient.Update(ctx, &existVSS)
	})
}
