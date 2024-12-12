package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ctfer-io/go-ctfd/api"
	"github.com/ctfer-io/terraform-provider-ctfd/provider/utils"
	"github.com/ctfer-io/terraform-provider-ctfd/provider/validators"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = (*challengeResource)(nil)
	_ resource.ResourceWithConfigure   = (*challengeResource)(nil)
	_ resource.ResourceWithImportState = (*challengeResource)(nil)
)

func NewChallengeResource() resource.Resource {
	return &challengeResource{}
}

type challengeResource struct {
	client *api.Client
}

type challengeResourceModel struct {
	ID             types.String                  `tfsdk:"id"`
	Name           types.String                  `tfsdk:"name"`
	Category       types.String                  `tfsdk:"category"`
	Description    types.String                  `tfsdk:"description"`
	ConnectionInfo types.String                  `tfsdk:"connection_info"`
	MaxAttempts    types.Int64                   `tfsdk:"max_attempts"`
	Function       types.String                  `tfsdk:"function"`
	Value          types.Int64                   `tfsdk:"value"`
	Decay          types.Int64                   `tfsdk:"decay"`
	Minimum        types.Int64                   `tfsdk:"minimum"`
	State          types.String                  `tfsdk:"state"`
	Type           types.String                  `tfsdk:"type"`
	Next           types.Int64                   `tfsdk:"next"`
	Requirements   *RequirementsSubresourceModel `tfsdk:"requirements"`
	Tags           []types.String                `tfsdk:"tags"`
	Topics         []types.String                `tfsdk:"topics"`
}

func (r *challengeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_challenge"
}

func (r *challengeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "CTFd is built around the Challenge resource, which contains all the attributes to define a part of the Capture The Flag event.\n\nThis provider builds a cleaner API on top of CTFd's one to improve its adoption and lifecycle management.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier of the challenge.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the challenge, displayed as it.",
				Required:            true,
			},
			"category": schema.StringAttribute{
				MarkdownDescription: "Category of the challenge that CTFd groups by on the web UI.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the challenge, consider using multiline descriptions for better style.",
				Required:            true,
			},
			"connection_info": schema.StringAttribute{
				MarkdownDescription: "Connection Information to connect to the challenge instance, useful for pwn, web and infrastructure pentests.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"max_attempts": schema.Int64Attribute{
				MarkdownDescription: "Maximum amount of attempts before being unable to flag the challenge.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"function": schema.StringAttribute{
				MarkdownDescription: "Decay function to define how the challenge value evolve through solves, either linear or logarithmic.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"value": schema.Int64Attribute{
				MarkdownDescription: "The value (points) of the challenge once solved. Internally, the provider will handle what target is legitimate depending on the `.type` value, i.e. either `value` for \"standard\" or `initial` for \"dynamic\".",
				Required:            true,
			},
			"decay": schema.Int64Attribute{
				MarkdownDescription: "The decay defines from each number of solves does the decay function triggers until reaching minimum. This function is defined by CTFd and could be configured through `.function`.",
				Optional:            true,
				Computed:            true,
			},
			"minimum": schema.Int64Attribute{
				MarkdownDescription: "The minimum points for a dynamic-score challenge to reach with the decay function. Once there, no solve could have more value.",
				Optional:            true,
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "State of the challenge, either hidden or visible.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("hidden"),
				Validators: []validator.String{
					validators.NewStringEnumValidator([]basetypes.StringValue{
						types.StringValue("hidden"),
						types.StringValue("visible"),
					}),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of the challenge defining its layout/behavior, either standard or dynamic (default).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("dynamic"),
				Validators: []validator.String{
					validators.NewStringEnumValidator([]basetypes.StringValue{
						types.StringValue("standard"),
						types.StringValue("dynamic"),
					}),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"next": schema.Int64Attribute{
				MarkdownDescription: "Suggestion for the end-user as next challenge to work on.",
				Optional:            true,
			},
			"requirements": schema.SingleNestedAttribute{
				MarkdownDescription: "List of required challenges that needs to get flagged before this one being accessible. Useful for skill-trees-like strategy CTF.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"behavior": schema.StringAttribute{
						MarkdownDescription: "Behavior if not unlocked, either hidden or anonymized.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("hidden"),
						Validators: []validator.String{
							validators.NewStringEnumValidator([]basetypes.StringValue{
								BehaviorHidden,
								BehaviorAnonymized,
							}),
						},
					},
					"prerequisites": schema.ListAttribute{
						MarkdownDescription: "List of the challenges ID.",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "List of challenge tags that will be displayed to the end-user. You could use them to give some quick insights of what a challenge involves.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(basetypes.NewListValueMust(types.StringType, []attr.Value{})),
			},
			"topics": schema.ListAttribute{
				MarkdownDescription: "List of challenge topics that are displayed to the administrators for maintenance and planification.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(basetypes.NewListValueMust(types.StringType, []attr.Value{})),
			},
		},
	}
}

func (r *challengeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *github.com/ctfer-io/go-ctfd/api.Client, got: %T. Please open an issue at https://github.com/ctfer-io/terraform-provider-ctfd", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *challengeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data challengeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration checks
	if data.Type.Equal(types.StringValue("dynamic")) {
		if data.Decay.IsNull() {
			resp.Diagnostics.AddError("Configuration error", "decay must be set for dynamic challenges")
		}
		if data.Minimum.IsNull() {
			resp.Diagnostics.AddError("Configuration error", "minimum must be set for dynamic challenges")
		}
		if data.Function.IsNull() || data.Function.IsUnknown() {
			data.Function = FunctionLogarithmic
		}
	} else {
		if data.Function.IsNull() || data.Function.IsUnknown() {
			data.Function = types.StringNull()
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Create Challenge
	reqs := (*api.Requirements)(nil)
	if data.Requirements != nil {
		preqs := make([]int, 0, len(data.Requirements.Prerequisites))
		for _, preq := range data.Requirements.Prerequisites {
			id, _ := strconv.Atoi(preq.ValueString())
			preqs = append(preqs, id)
		}
		reqs = &api.Requirements{
			Anonymize:     GetAnon(data.Requirements.Behavior),
			Prerequisites: preqs,
		}
	}
	res, err := r.client.PostChallenges(&api.PostChallengesParams{
		Name:           data.Name.ValueString(),
		Category:       data.Category.ValueString(),
		Description:    data.Description.ValueString(),
		ConnectionInfo: data.ConnectionInfo.ValueStringPointer(),
		MaxAttempts:    utils.ToInt(data.MaxAttempts),
		Function:       data.Function.ValueStringPointer(),
		Value:          int(data.Value.ValueInt64()),
		Initial:        utils.ToIntOnDynamic(data.Value, data.Type),
		Decay:          utils.ToIntOnDynamic(data.Decay, data.Type),
		Minimum:        utils.ToIntOnDynamic(data.Minimum, data.Type),
		State:          data.State.ValueString(),
		Type:           data.Type.ValueString(),
		NextID:         utils.ToInt(data.Next),
		Requirements:   reqs,
	}, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create challenge, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "created a challenge")

	// Save computed attributes in state
	data.ID = types.StringValue(strconv.Itoa(res.ID))
	data.Decay = utils.ToTFInt64(res.Decay)
	data.Minimum = utils.ToTFInt64(res.Minimum)

	// Create tags
	challTags := make([]types.String, 0, len(data.Tags))
	for _, tag := range data.Tags {
		_, err := r.client.PostTags(&api.PostTagsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Value:     tag.ValueString(),
		}, api.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create tags, got error: %s", err),
			)
			return
		}
		challTags = append(challTags, tag)
	}
	if data.Tags != nil {
		data.Tags = challTags
	}

	// Create topics
	challTopics := make([]types.String, 0, len(data.Topics))
	for _, topic := range data.Topics {
		_, err := r.client.PostTopics(&api.PostTopicsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Type:      "challenge",
			Value:     topic.ValueString(),
		}, api.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create topic, got error: %s", err),
			)
			return
		}
		challTopics = append(challTopics, topic)
	}
	if data.Topics != nil {
		data.Topics = challTopics
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *challengeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data challengeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Read(ctx, r.client, resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *challengeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data challengeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var dataState challengeResourceModel
	req.State.Get(ctx, &dataState)

	// Configuration checks
	if data.Type.Equal(types.StringValue("dynamic")) {
		if data.Decay.IsNull() {
			resp.Diagnostics.AddError("Configuration error", "decay must be set for dynamic challenges")
		}
		if data.Minimum.IsNull() {
			resp.Diagnostics.AddError("Configuration error", "minimum must be set for dynamic challenges")
		}
		if data.Function.IsNull() {
			data.Function = FunctionLogarithmic
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Patch direct attributes
	reqs := (*api.Requirements)(nil)
	if data.Requirements != nil {
		preqs := make([]int, 0, len(data.Requirements.Prerequisites))
		for _, preq := range data.Requirements.Prerequisites {
			id, _ := strconv.Atoi(preq.ValueString())
			preqs = append(preqs, id)
		}
		reqs = &api.Requirements{
			Anonymize:     GetAnon(data.Requirements.Behavior),
			Prerequisites: preqs,
		}
	}
	_, err := r.client.PatchChallenge(utils.Atoi(data.ID.ValueString()), &api.PatchChallengeParams{
		Name:           data.Name.ValueString(),
		Category:       data.Category.ValueString(),
		Description:    data.Description.ValueString(),
		ConnectionInfo: data.ConnectionInfo.ValueStringPointer(),
		MaxAttempts:    utils.ToInt(data.MaxAttempts),
		Function:       data.Function.ValueStringPointer(),
		Value:          utils.ToInt(data.Value),
		Initial:        utils.ToInt(data.Value),
		Decay:          utils.ToInt(data.Decay),
		Minimum:        utils.ToInt(data.Minimum),
		State:          data.State.ValueString(),
		NextID:         utils.ToInt(data.Next),
		Requirements:   reqs,
	}, api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update challenge, got error: %s", err),
		)
		return
	}

	// Update its tags (drop them all, create new ones)
	challTags, err := r.client.GetChallengeTags(utils.Atoi(data.ID.ValueString()), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get all tags of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	for _, tag := range challTags {
		if err := r.client.DeleteTag(strconv.Itoa(tag.ID), api.WithContext(ctx)); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to delete tag %d of challenge %s, got error: %s", tag.ID, data.ID.ValueString(), err),
			)
			return
		}
	}
	tags := make([]types.String, 0, len(data.Tags))
	for _, tag := range data.Tags {
		_, err := r.client.PostTags(&api.PostTagsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Value:     tag.ValueString(),
		}, api.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create tag of challenge %s, got error: %s", data.ID.ValueString(), err),
			)
			return
		}
		tags = append(tags, tag)
	}
	if data.Tags != nil {
		data.Tags = tags
	}

	// Update its topics (drop them all, create new ones)
	challTopics, err := r.client.GetChallengeTopics(utils.Atoi(data.ID.ValueString()), api.WithContext(ctx))
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get all topics of challenge %s, got error: %s", data.ID.ValueString(), err),
		)
		return
	}
	for _, topic := range challTopics {
		if err := r.client.DeleteTopic(&api.DeleteTopicArgs{
			ID:   strconv.Itoa(topic.ID),
			Type: "challenge",
		}, api.WithContext(ctx)); err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to delete topic %d of challenge %s, got error: %s", topic.ID, data.ID.ValueString(), err),
			)
			return
		}
	}
	topics := make([]types.String, 0, len(data.Topics))
	for _, topic := range data.Topics {
		_, err := r.client.PostTopics(&api.PostTopicsParams{
			Challenge: utils.Atoi(data.ID.ValueString()),
			Type:      "challenge",
			Value:     topic.ValueString(),
		}, api.WithContext(ctx))
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to create topic of challenge %s, got error: %s", data.ID.ValueString(), err),
			)
			return
		}
		topics = append(topics, topic)
	}
	if data.Topics != nil {
		data.Topics = topics
	}

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *challengeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data challengeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteChallenge(utils.Atoi(data.ID.ValueString()), api.WithContext(ctx)); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete challenge, got error: %s", err))
		return
	}

	// ... don't need to delete nested objects, this is handled by CTFd
}

func (r *challengeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Automatically call r.Read
}

//
// Starting from this are helper or types-specific code related to the ctfd_challenge resource
//

func (chall *challengeResourceModel) Read(ctx context.Context, client *api.Client, diags diag.Diagnostics) {
	res, err := client.GetChallenge(utils.Atoi(chall.ID.ValueString()), api.WithContext(ctx))
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read challenge %s, got error: %s", chall.ID.ValueString(), err))
		return
	}
	chall.Name = types.StringValue(res.Name)
	chall.Category = types.StringValue(res.Category)
	chall.Description = types.StringValue(res.Description)
	chall.ConnectionInfo = utils.ToTFString(res.ConnectionInfo)
	chall.MaxAttempts = utils.ToTFInt64(res.MaxAttempts)
	chall.Function = utils.ToTFString(res.Function)
	chall.Decay = utils.ToTFInt64(res.Decay)
	chall.Minimum = utils.ToTFInt64(res.Minimum)
	chall.State = types.StringValue(res.State)
	chall.Type = types.StringValue(res.Type)
	chall.Next = utils.ToTFInt64(res.NextID)

	switch res.Type {
	case "standard":
		chall.Value = types.Int64Value(int64(res.Value))
	case "dynamic":
		chall.Value = utils.ToTFInt64(res.Initial)
	}

	id := utils.Atoi(chall.ID.ValueString())

	// Get subresources
	// => Requirements
	resReqs, err := client.GetChallengeRequirements(id, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d requirements, got error: %s", id, err),
		)
		return
	}
	reqs := (*RequirementsSubresourceModel)(nil)
	if resReqs != nil {
		challPreqs := make([]types.String, 0, len(resReqs.Prerequisites))
		for _, req := range resReqs.Prerequisites {
			challPreqs = append(challPreqs, types.StringValue(strconv.Itoa(req)))
		}
		reqs = &RequirementsSubresourceModel{
			Behavior:      FromAnon(resReqs.Anonymize),
			Prerequisites: challPreqs,
		}
	}
	chall.Requirements = reqs

	// => Tags
	resTags, err := client.GetChallengeTags(id, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d tags, got error: %s", id, err),
		)
		return
	}
	chall.Tags = make([]basetypes.StringValue, 0, len(resTags))
	for _, tag := range resTags {
		chall.Tags = append(chall.Tags, types.StringValue(tag.Value))
	}

	// => Topics
	resTopics, err := client.GetChallengeTopics(id, api.WithContext(ctx))
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read challenge %d topics, got error: %s", id, err),
		)
		return
	}
	chall.Topics = make([]basetypes.StringValue, 0, len(resTopics))
	for _, topic := range resTopics {
		chall.Topics = append(chall.Topics, types.StringValue(topic.Value))
	}
}

var (
	BehaviorHidden     = types.StringValue("hidden")
	BehaviorAnonymized = types.StringValue("anonymized")

	FunctionLinear      = types.StringValue("linear")
	FunctionLogarithmic = types.StringValue("logarithmic")
)

type RequirementsSubresourceModel struct {
	Behavior      types.String   `tfsdk:"behavior"`
	Prerequisites []types.String `tfsdk:"prerequisites"`
}

func GetAnon(str types.String) *bool {
	switch {
	case str.Equal(BehaviorHidden):
		return nil
	case str.Equal(BehaviorAnonymized):
		return utils.Ptr(true)
	}
	panic("invalid anonymization value: " + str.ValueString())
}

func FromAnon(b *bool) types.String {
	if b == nil {
		return BehaviorHidden
	}
	if *b {
		return BehaviorAnonymized
	}
	panic("invalid anonymization value, got boolean false")
}
