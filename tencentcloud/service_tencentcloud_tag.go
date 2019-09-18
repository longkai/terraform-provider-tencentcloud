package tencentcloud

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/resource"
	tag "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tag/v20180813"
	"github.com/terraform-providers/terraform-provider-tencentcloud/tencentcloud/connectivity"
	"github.com/terraform-providers/terraform-provider-tencentcloud/tencentcloud/ratelimit"
)

type TagService struct {
	client *connectivity.TencentCloudClient
}

func (me *TagService) ModifyTags(ctx context.Context, resource string, replaceTags map[string]string, deleteKeys []string) error {
	logId := getLogId(ctx)
	request := tag.NewModifyResourceTagsRequest()
	request.Resource = &resource
	if len(replaceTags) > 0 {
		request.ReplaceTags = make([]*tag.Tag, 0, len(replaceTags))
		for k, v := range replaceTags {
			key := k
			value := v
			replaceTag := &tag.Tag{
				TagKey:   &key,
				TagValue: &value,
			}
			request.ReplaceTags = append(request.ReplaceTags, replaceTag)
		}
	}
	if len(deleteKeys) > 0 {
		request.DeleteTags = make([]*tag.TagKeyObject, 0, len(deleteKeys))
		for _, v := range deleteKeys {
			key := v
			deleteKey := &tag.TagKeyObject{
				TagKey: &key,
			}
			request.DeleteTags = append(request.DeleteTags, deleteKey)
		}
	}

	ratelimit.Check(request.GetAction())
	response, err := me.client.UseTagClient().ModifyResourceTags(request)
	if err != nil {
		log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]\n",
			logId, request.GetAction(), request.ToJsonString(), err.Error())
		return err
	}
	log.Printf("[DEBUG]%s api[%s] success, request body [%s], response body [%s]\n",
		logId, request.GetAction(), request.ToJsonString(), response.ToJsonString())

	return nil
}

func (me *TagService) DescribeResourceTags(ctx context.Context, serviceType, resourceType, region, resourceId string) (tags map[string]string, err error) {
	logId := getLogId(ctx)
	client := me.client.UseTagClient()

	request := tag.NewDescribeResourceTagsByResourceIdsRequest()
	request.ServiceType = &serviceType
	request.ResourcePrefix = &resourceType
	request.ResourceRegion = &region
	request.ResourceIds = []*string{&resourceId}
	request.Limit = intToPointer(20)

	var offset uint64
	request.Offset = &offset

	// for run loop at least once
	count := 20
	for count == 20 {
		if err := resource.Retry(readRetryTimeout, func() *resource.RetryError {
			ratelimit.Check(request.GetAction())

			response, err := client.DescribeResourceTagsByResourceIds(request)
			if err != nil {
				count = 0

				log.Printf("[CRITAL]%s api[%s] fail, request body [%s], reason[%s]",
					logId, request.GetAction(), request.ToJsonString(), err.Error())
				return retryError(err)
			}

			allTags := response.Response.Tags
			count = len(allTags)

			for _, t := range allTags {
				if nilFields := CheckNil(t, map[string]string{
					"ResourceId": "resource id",
					"TagKey":     "tag key",
					"TagValue":   "tag value",
				}); len(nilFields) > 0 {
					err := fmt.Errorf("api[%s] tag %v are nil", request.GetAction(), nilFields)
					log.Printf("[CRITAL]%s %v", logId, err)
					return resource.NonRetryableError(err)
				}

				if *t.ResourceId != resourceId {
					continue
				}
				if tags == nil {
					tags = make(map[string]string)
				}

				tags[*t.TagKey] = *t.TagValue
			}

			return nil
		}); err != nil {
			log.Printf("[CRITAL]%s describe resource %s tag failed, reason: %v", logId, serviceType+":"+resourceType, err)
			return nil, err
		}

		offset += uint64(count)
	}

	return
}

func diffTags(oldTags, newTags map[string]interface{}) (replaceTags map[string]string, deleteTags []string) {
	replaceTags = make(map[string]string)
	deleteTags = make([]string, 0)
	for k, v := range newTags {
		_, ok := oldTags[k]
		if !ok || oldTags[k].(string) != v.(string) {
			replaceTags[k] = v.(string)
		}
	}
	for k := range oldTags {
		_, ok := newTags[k]
		if !ok {
			deleteTags = append(deleteTags, k)
		}
	}
	return
}

// COS and CAS should use buildResourceNameWithUid
func buildResourceNameWithUin(serviceType, resourceType, id, region string) string {
	return fmt.Sprintf("qcs::%s:%s:uin/:%s/%s", serviceType, region, resourceType, id)
}

func buildResourceNameWithUid(serviceType, resourceType, id, region string) string {
	return fmt.Sprintf("qcs::%s:%s:uid/:%s/%s", serviceType, region, resourceType, id)
}
