package appmesh

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appmesh_virtual_router")
func ResourceVirtualRouter() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceVirtualRouterCreate,
		ReadWithoutTimeout:   resourceVirtualRouterRead,
		UpdateWithoutTimeout: resourceVirtualRouterUpdate,
		DeleteWithoutTimeout: resourceVirtualRouterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceVirtualRouterImport,
		},

		SchemaVersion: 1,
		MigrateState:  resourceVirtualRouterMigrateState,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mesh_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"mesh_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"spec":            resourceVirtualRouterSpecSchema(),
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVirtualRouterSpecSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"listener": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"port_mapping": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"port": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IsPortNumber,
										},
										"protocol": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice(appmesh.PortProtocol_Values(), false),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceVirtualRouterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(ctx, d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &appmesh.CreateVirtualRouterInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		Spec:              expandVirtualRouterSpec(d.Get("spec").([]interface{})),
		Tags:              Tags(tags.IgnoreAWS()),
		VirtualRouterName: aws.String(name),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	output, err := conn.CreateVirtualRouterWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh Virtual Router (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.VirtualRouter.Metadata.Uid))

	return append(diags, resourceVirtualRouterRead(ctx, d, meta)...)
}

func resourceVirtualRouterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindVirtualRouterByThreePartKey(ctx, conn, d.Get("mesh_name").(string), d.Get("mesh_owner").(string), d.Get("name").(string))
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Mesh Virtual Router (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Router (%s): %s", d.Id(), err)
	}

	vr := outputRaw.(*appmesh.VirtualRouterData)

	arn := aws.StringValue(vr.Metadata.Arn)
	d.Set("arn", arn)
	d.Set("created_date", vr.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", vr.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", vr.MeshName)
	d.Set("mesh_owner", vr.Metadata.MeshOwner)
	d.Set("name", vr.VirtualRouterName)
	d.Set("resource_owner", vr.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenVirtualRouterSpec(vr.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh Virtual Router (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceVirtualRouterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()

	if d.HasChange("spec") {
		input := &appmesh.UpdateVirtualRouterInput{
			MeshName:          aws.String(d.Get("mesh_name").(string)),
			Spec:              expandVirtualRouterSpec(d.Get("spec").([]interface{})),
			VirtualRouterName: aws.String(d.Get("name").(string)),
		}

		if v, ok := d.GetOk("mesh_owner"); ok {
			input.MeshOwner = aws.String(v.(string))
		}

		_, err := conn.UpdateVirtualRouterWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Virtual Router (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		arn := d.Get("arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Virtual Router (%s) tags: %s", arn, err)
		}
	}

	return append(diags, resourceVirtualRouterRead(ctx, d, meta)...)
}

func resourceVirtualRouterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()

	log.Printf("[DEBUG] Deleting App Mesh Virtual Router: %s", d.Id())
	input := &appmesh.DeleteVirtualRouterInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		VirtualRouterName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	_, err := conn.DeleteVirtualRouterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh Virtual Router (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceVirtualRouterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'mesh-name/virtual-router-name'", d.Id())
	}

	conn := meta.(*conns.AWSClient).AppMeshConn()
	meshName := parts[0]
	name := parts[1]

	vr, err := FindVirtualRouterByThreePartKey(ctx, conn, meshName, "", name)

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(vr.Metadata.Uid))
	d.Set("mesh_name", vr.MeshName)
	d.Set("name", vr.VirtualRouterName)

	return []*schema.ResourceData{d}, nil
}

func FindVirtualRouterByThreePartKey(ctx context.Context, conn *appmesh.AppMesh, meshName, meshOwner, name string) (*appmesh.VirtualRouterData, error) {
	input := &appmesh.DescribeVirtualRouterInput{
		MeshName:          aws.String(meshName),
		VirtualRouterName: aws.String(name),
	}
	if meshOwner != "" {
		input.MeshOwner = aws.String(meshOwner)
	}

	output, err := findVirtualRouter(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status.Status); status == appmesh.VirtualRouterStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findVirtualRouter(ctx context.Context, conn *appmesh.AppMesh, input *appmesh.DescribeVirtualRouterInput) (*appmesh.VirtualRouterData, error) {
	output, err := conn.DescribeVirtualRouterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VirtualRouter == nil || output.VirtualRouter.Metadata == nil || output.VirtualRouter.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VirtualRouter, nil
}
