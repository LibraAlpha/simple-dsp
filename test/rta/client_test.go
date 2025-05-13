package rta_test

import (
    "context"
    "testing"
    "time"
    
    "simple-dsp/internal/rta"
    "github.com/stretchr/testify/assert"
)

func TestClient_SingleQuery(t *testing.T) {
    client := rta.NewClient("test_app_key", "test_app_secret")
    
    tests := []struct {
        name    string
        req     *rta.SingleRequest
        wantErr bool
    }{
        {
            name: "正常Android IMEI请求",
            req: &rta.SingleRequest{
                Channel:            "test_channel",
                AdvertisingSpaceID: "test_ad_space",
                IMEI:              "123456789012345",
                OS:                "0",
            },
            wantErr: false,
        },
        {
            name: "正常iOS IDFA请求",
            req: &rta.SingleRequest{
                Channel:            "test_channel",
                AdvertisingSpaceID: "test_ad_space",
                IDFA:              "ABCDEF01-2345-6789-ABCD-EF0123456789",
                OS:                "1",
            },
            wantErr: false,
        },
        {
            name: "缺少渠道ID",
            req: &rta.SingleRequest{
                AdvertisingSpaceID: "test_ad_space",
                IMEI:              "123456789012345",
            },
            wantErr: true,
        },
        {
            name: "缺少广告位ID",
            req: &rta.SingleRequest{
                Channel: "test_channel",
                IMEI:   "123456789012345",
            },
            wantErr: true,
        },
        {
            name: "缺少设备ID",
            req: &rta.SingleRequest{
                Channel:            "test_channel",
                AdvertisingSpaceID: "test_ad_space",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(context.Background(), time.Second)
            defer cancel()
            
            resp, err := client.SingleQuery(ctx, tt.req)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.NotNil(t, resp)
        })
    }
}

func TestClient_BatchQuery(t *testing.T) {
    client := rta.NewClient("test_app_key", "test_app_secret")
    
    tests := []struct {
        name    string
        req     *rta.BatchRequest
        wantErr bool
    }{
        {
            name: "正常批量IMEI请求",
            req: &rta.BatchRequest{
                Channel:            "test_channel",
                AdvertisingSpaceID: "test_ad_space",
                IMEIMD5List:       "abc123,def456,ghi789",
            },
            wantErr: false,
        },
        {
            name: "正常批量IDFA请求",
            req: &rta.BatchRequest{
                Channel:            "test_channel",
                AdvertisingSpaceID: "test_ad_space",
                IDFAMD5List:       "abc123,def456,ghi789",
            },
            wantErr: false,
        },
        {
            name: "超过设备ID数量限制",
            req: &rta.BatchRequest{
                Channel:            "test_channel",
                AdvertisingSpaceID: "test_ad_space",
                IMEIMD5List:       "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21",
            },
            wantErr: true,
        },
        {
            name: "缺少渠道ID",
            req: &rta.BatchRequest{
                AdvertisingSpaceID: "test_ad_space",
                IMEIMD5List:       "abc123,def456",
            },
            wantErr: true,
        },
        {
            name: "缺少广告位ID",
            req: &rta.BatchRequest{
                Channel:     "test_channel",
                IMEIMD5List: "abc123,def456",
            },
            wantErr: true,
        },
        {
            name: "缺少设备ID列表",
            req: &rta.BatchRequest{
                Channel:            "test_channel",
                AdvertisingSpaceID: "test_ad_space",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(context.Background(), time.Second)
            defer cancel()
            
            resp, err := client.BatchQuery(ctx, tt.req)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.NotNil(t, resp)
        })
    }
} 