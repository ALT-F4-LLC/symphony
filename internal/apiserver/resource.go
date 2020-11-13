package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/state"
	"github.com/google/uuid"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Resources : defines handlers for resources
type Resources struct {
	EtcdEndpoints []string
}

func (r *Resources) getClusters() ([]*api.Cluster, error) {
	results, err := r.getStateByKey("/Cluster", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	clusters := make([]*api.Cluster, 0)

	for _, ev := range results.Kvs {
		var c *api.Cluster

		err := json.Unmarshal(ev.Value, &c)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		clusters = append(clusters, c)
	}

	return clusters, nil
}

func (r *Resources) getClusterByID(clusterID uuid.UUID) (*api.Cluster, error) {
	clusterKey := fmt.Sprintf("/Cluster/%s", clusterID.String())

	results, err := r.getStateByKey(clusterKey)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var cluster *api.Cluster

	for i, ev := range results.Kvs {

		if i == 0 {
			err := json.Unmarshal(ev.Value, &cluster)

			if err != nil {
				st := status.New(codes.Internal, err.Error())

				return nil, st.Err()
			}
		}
	}

	return cluster, nil
}

func (r *Resources) getLogicalVolumes() ([]*api.LogicalVolume, error) {
	results, err := r.getStateByKey("/LogicalVolume", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	lvs := make([]*api.LogicalVolume, 0)

	for _, ev := range results.Kvs {
		var lv *api.LogicalVolume

		err := json.Unmarshal(ev.Value, &lv)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		lvs = append(lvs, lv)
	}

	return lvs, nil
}

func (r *Resources) getLogicalVolumeByID(id uuid.UUID) (*api.LogicalVolume, error) {
	etcdKey := fmt.Sprintf("/LogicalVolume/%s", id.String())

	results, err := r.getStateByKey(etcdKey)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var lv *api.LogicalVolume

	for _, ev := range results.Kvs {
		var v *api.LogicalVolume

		err := json.Unmarshal(ev.Value, &v)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		if v.ID == id.String() {
			lv = v
		}
	}

	return lv, nil
}

func (r *Resources) getPhysicalVolumes() ([]*api.PhysicalVolume, error) {
	results, err := r.getStateByKey("/PhysicalVolume", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	pvs := make([]*api.PhysicalVolume, 0)

	for _, ev := range results.Kvs {
		var s *api.PhysicalVolume

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		pvs = append(pvs, s)
	}

	return pvs, nil
}

func (m *Resources) getPhysicalVolumeByID(id uuid.UUID) (*api.PhysicalVolume, error) {
	etcdKey := fmt.Sprintf("/physicalvolume/%s", id.String())

	results, err := m.getStateByKey(etcdKey)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var volume *api.PhysicalVolume

	for _, ev := range results.Kvs {
		var s *api.PhysicalVolume

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		if s.ID == id.String() {
			volume = s
		}
	}

	return volume, nil
}

func (r *Resources) getServices() ([]*api.Service, error) {
	results, err := r.getStateByKey("/Service", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	services := make([]*api.Service, 0)

	for _, ev := range results.Kvs {
		var s *api.Service

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		services = append(services, s)
	}

	return services, nil
}

func (r *Resources) getServiceByID(id uuid.UUID) (*api.Service, error) {
	etcdKey := fmt.Sprintf("/Service/%s", id.String())

	results, err := r.getStateByKey(etcdKey)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var service *api.Service

	for _, ev := range results.Kvs {
		var s *api.Service

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		if s.ID == id.String() {
			service = s
		}
	}

	return service, nil
}

func (r *Resources) getStateByKey(key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), EtcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: EtcdDialTimeout,
		Endpoints:   r.EtcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return nil, err
	}

	defer client.Close()

	results, err := client.Get(ctx, key, opts...)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (m *Resources) getVolumeGroups() ([]*api.VolumeGroup, error) {
	results, err := m.getStateByKey("/VolumeGroup", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	vgs := make([]*api.VolumeGroup, 0)

	for _, ev := range results.Kvs {
		var s *api.VolumeGroup

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		vgs = append(vgs, s)
	}

	return vgs, nil
}

func (r *Resources) getVolumeGroupByID(id uuid.UUID) (*api.VolumeGroup, error) {
	etcdKey := fmt.Sprintf("/VolumeGroup/%s", id.String())

	results, err := r.getStateByKey(etcdKey)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var volumeGroup *api.VolumeGroup

	for _, ev := range results.Kvs {
		var vg *api.VolumeGroup

		err := json.Unmarshal(ev.Value, &vg)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		if vg.ID == id.String() {
			volumeGroup = vg
		}
	}

	return volumeGroup, nil
}

func (r *Resources) removeResource(key string) error {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return st.Err()
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	_, res := etcd.Delete(ctx, key)

	if res != nil {
		st := status.New(codes.Internal, res.Error())

		return st.Err()
	}

	return nil
}

func (r *Resources) saveCluster(cluster *api.Cluster) error {
	ctx, cancel := context.WithTimeout(context.Background(), EtcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: EtcdDialTimeout,
		Endpoints:   r.EtcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return err
	}

	defer client.Close()

	clusterJSON, err := json.Marshal(cluster)

	if err != nil {
		return err
	}

	clusterKey := fmt.Sprintf("/Cluster/%s", cluster.ID)

	_, putErr := client.Put(ctx, clusterKey, string(clusterJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (r *Resources) saveLogicalVolume(lv *api.LogicalVolume) error {
	ctx, cancel := context.WithTimeout(context.Background(), EtcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: EtcdDialTimeout,
		Endpoints:   r.EtcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return err
	}

	defer client.Close()

	lvJSON, err := json.Marshal(lv)

	if err != nil {
		return err
	}

	lvKey := fmt.Sprintf("/LogicalVolume/%s", lv.ID)

	_, putErr := client.Put(ctx, lvKey, string(lvJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (r *Resources) savePhysicalVolume(pv *api.PhysicalVolume) error {
	ctx, cancel := context.WithTimeout(context.Background(), EtcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: EtcdDialTimeout,
		Endpoints:   r.EtcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return err
	}

	defer client.Close()

	pvJSON, err := json.Marshal(pv)

	if err != nil {
		return err
	}

	volumeKey := fmt.Sprintf("/PhysicalVolume/%s", pv.ID)

	_, putErr := client.Put(ctx, volumeKey, string(pvJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (r *Resources) saveVolumeGroup(vg *api.VolumeGroup) error {
	ctx, cancel := context.WithTimeout(context.Background(), EtcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: EtcdDialTimeout,
		Endpoints:   r.EtcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return err
	}

	defer client.Close()

	vgJSON, err := json.Marshal(vg)

	if err != nil {
		return err
	}

	volumeKey := fmt.Sprintf("/VolumeGroup/%s", vg.ID)

	_, putErr := client.Put(ctx, volumeKey, string(vgJSON))

	if putErr != nil {
		return err
	}

	return nil
}

func (m *Resources) saveService(service *api.Service) error {
	ctx, cancel := context.WithTimeout(context.Background(), EtcdDialTimeout)

	defer cancel()

	config := clientv3.Config{
		DialTimeout: EtcdDialTimeout,
		Endpoints:   m.EtcdEndpoints,
	}

	client, err := state.NewClient(config)

	if err != nil {
		return err
	}

	defer client.Close()

	serviceKey := fmt.Sprintf("/Service/%s", service.ID)

	serviceJSON, err := json.Marshal(service)

	if err != nil {
		return err
	}

	_, putErr := client.Put(ctx, serviceKey, string(serviceJSON))

	if putErr != nil {
		return err
	}

	return nil
}
