package helm

import (
	"k8s.io/helm/pkg/helm"
	rspb "k8s.io/helm/pkg/proto/hapi/release"
)

func installOpts(release string, wait bool, vals []byte) []helm.InstallOption {
	return []helm.InstallOption{
		helm.InstallReuseName(true),
		helm.InstallWait(wait),
		helm.ReleaseName(release),
		helm.ValueOverrides(vals),
	}
}

func updateOpts(release string, vals []byte) []helm.UpdateOption {
	return []helm.UpdateOption{
		helm.UpdateValueOverrides(vals),
		helm.UpgradeWait(true),
	}
}

func deleteOpts() []helm.DeleteOption {
	return []helm.DeleteOption{
		helm.DeletePurge(true),
	}
}

func listOpts(release string) []helm.ReleaseListOption {
	return []helm.ReleaseListOption{
		helm.ReleaseListFilter(release),
		helm.ReleaseListStatuses([]rspb.Status_Code{
			rspb.Status_DELETING,
			rspb.Status_DEPLOYED,
			rspb.Status_FAILED,
			rspb.Status_PENDING_INSTALL,
			rspb.Status_PENDING_ROLLBACK,
			rspb.Status_PENDING_UPGRADE,
			rspb.Status_UNKNOWN,
		}),
	}
}
