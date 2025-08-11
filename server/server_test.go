package server

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/sinnlos-ffff/tsdb-lite/database"
	pb "github.com/sinnlos-ffff/tsdb-lite/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func setupTestServer(t *testing.T) (pb.TsdbLiteClient, *Server) {
	// Create a new server
	server := NewServer(&Config{
		CompactionInterval: time.Minute,
	})

	// Create a listener on a random port
	lis, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	// Start the server in a goroutine
	go func() {
		if err := server.grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Create a client connection to the server
	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	// Create a client
	client := pb.NewTsdbLiteClient(conn)

	t.Cleanup(func() {
		server.grpcServer.Stop()
		conn.Close()
	})

	return client, server
}

func TestServer_Integration_GRPC(t *testing.T) {
	client, server := setupTestServer(t)

	t.Run("CreateTimeSeries", func(t *testing.T) {
		metric := "cpu_usage"
		tags := map[string]string{"host": "server1", "region": "us-west"}

		_, err := client.CreateTimeSeries(context.Background(), &pb.CreateTimeSeriesRequest{
			Metric: metric,
			Tags:   tags,
		})
		require.NoError(t, err)

		// Verify the time series was actually created in the database
		key := database.GenerateKey(metric, tags)
		ts, ok := server.Db.GetShard(key).Series[key]
		assert.True(t, ok)
		assert.Equal(t, metric, ts.Metric)
		assert.Equal(t, tags, ts.Tags)
	})

	t.Run("AddPoint", func(t *testing.T) {
		metric := "disk_usage"
		tags := map[string]string{"host": "server3", "mount": "/var"}
		timestamp := time.Now().Unix()
		value := 75.5

		// First create the time series
		_, err := client.CreateTimeSeries(context.Background(), &pb.CreateTimeSeriesRequest{
			Metric: metric,
			Tags:   tags,
		})
		require.NoError(t, err)

		// Now add a point
		_, err = client.AddPoint(context.Background(), &pb.AddPointRequest{
			Metric:    metric,
			Timestamp: timestamp,
			Value:     value,
			Tags:      tags,
		})
		require.NoError(t, err)

		// Verify the point was added
		key := database.GenerateKey(metric, tags)
		ts, ok := server.Db.GetShard(key).Series[key]
		require.True(t, ok)
		require.Len(t, ts.Chunks, 1)
		require.Len(t, ts.Chunks[0].Points, 1)
		assert.Equal(t, timestamp, ts.Chunks[0].Points[0].Timestamp)
		assert.Equal(t, value, ts.Chunks[0].Points[0].Value)
	})

	t.Run("GetRange", func(t *testing.T) {
		metric := "temperature"
		tags := map[string]string{"sensor": "A1", "location": "datacenter"}

		// Create time series
		_, err := client.CreateTimeSeries(context.Background(), &pb.CreateTimeSeriesRequest{
			Metric: metric,
			Tags:   tags,
		})
		require.NoError(t, err)

		// Add multiple points
		points := []struct {
			timestamp int64
			value     float64
		}{
			{1000, 20.5},
			{2000, 21.0},
			{3000, 19.8},
			{4000, 22.3},
		}

		for _, point := range points {
			_, err = client.AddPoint(context.Background(), &pb.AddPointRequest{
				Metric:    metric,
				Timestamp: point.timestamp,
				Value:     point.value,
				Tags:      tags,
			})
			require.NoError(t, err)
		}

		// Query range
		resp, err := client.GetRange(context.Background(), &pb.GetRangeRequest{
			Metric: metric,
			Tags:   tags,
			Start:  0,
			End:    5000,
		})
		require.NoError(t, err)

		assert.Len(t, resp.Points, 4)

		// Verify points are returned in order
		for i, expectedPoint := range points {
			assert.Equal(t, expectedPoint.timestamp, resp.Points[i].Timestamp)
			assert.Equal(t, expectedPoint.value, resp.Points[i].Value)
		}
	})

	t.Run("GetRange - partial range", func(t *testing.T) {
		metric := "pressure"
		tags := map[string]string{"gauge": "B2"}

		// Create time series and add points (similar to above)
		_, err := client.CreateTimeSeries(context.Background(), &pb.CreateTimeSeriesRequest{
			Metric: metric,
			Tags:   tags,
		})
		require.NoError(t, err)

		// Add points
		timestamps := []int64{1000, 2000, 3000, 4000}
		values := []float64{100.0, 105.0, 95.0, 110.0}

		for i, ts := range timestamps {
			_, err = client.AddPoint(context.Background(), &pb.AddPointRequest{
				Metric:    metric,
				Timestamp: ts,
				Value:     values[i],
				Tags:      tags,
			})
			require.NoError(t, err)
		}

		// Query partial range (should return only middle two points)
		resp, err := client.GetRange(context.Background(), &pb.GetRangeRequest{
			Metric: metric,
			Tags:   tags,
			Start:  1500, // Between first and second point
			End:    3500, // Between third and fourth point
		})
		require.NoError(t, err)

		// Should return only the middle two points
		assert.Len(t, resp.Points, 2)
		assert.Equal(t, int64(2000), resp.Points[0].Timestamp)
		assert.Equal(t, 105.0, resp.Points[0].Value)
		assert.Equal(t, int64(3000), resp.Points[1].Timestamp)
		assert.Equal(t, 95.0, resp.Points[1].Value)
	})
}
