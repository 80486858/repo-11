package sentry

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jianyuan/go-sentry/v2/sentry"
)

func resourceSentryFilter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSentryFilterCreate,
		ReadContext:   resourceSentryFilterRead,
		UpdateContext: resourceSentryFilterUpdate,
		DeleteContext: resourceSentryFilterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: importOrganizationProjectAndID,
		},

		Schema: map[string]*schema.Schema{
			"organization": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug of the organization the project belongs to",
			},
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The slug of the project to create the plugin for",
			},
			"browser_extension": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether to filter out events from browser extension",
			},
			"legacy_browsers": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Events from these legacy browsers will be ignored",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceSentryFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Same as Update
	return resourceSentryFilterUpdate(ctx, d, meta)
}

func resourceSentryFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*sentry.Client)
	org := d.Get("organization").(string)
	project := d.Get("project").(string)

	tflog.Debug(ctx, "Reading Sentry filter config", map[string]interface{}{"org": org, "project": project})
	filterConfig, resp, err := client.ProjectFilters.GetFilterConfig(ctx, org, project)
	if found, err := checkClientGet(resp, err, d); !found {
		return diag.FromErr(err)
	}
	tflog.Trace(ctx, "Read Sentry filter config", map[string]interface{}{"filterConfig": filterConfig})

	d.SetId(fmt.Sprintf("%s-%s_filter", org, project))
	d.Set("browser_extension", filterConfig.BrowserExtension)
	d.Set("legacy_browsers", filterConfig.LegacyBrowsers)

	return nil
}

func resourceSentryFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*sentry.Client)

	org := d.Get("organization").(string)
	project := d.Get("project").(string)

	browserExtension := d.Get("browser_extension").(bool)
	inputLegacyBrowsers := d.Get("legacy_browsers").(*schema.Set).List()
	legacyBrowsers := expandStringList(inputLegacyBrowsers)

	tflog.Debug(ctx, "Updating Sentry filters browser extensions and legacy browser", map[string]interface{}{"org": org, "project": project})
	_, err := client.ProjectFilters.UpdateBrowserExtensions(ctx, org, project, browserExtension)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.ProjectFilters.UpdateLegacyBrowser(ctx, org, project, legacyBrowsers)
	if err != nil {
		return diag.FromErr(err)
	}
	tflog.Debug(ctx, "Updated Sentry filters browser extensions and legacy browser", map[string]interface{}{"org": org, "project": project})

	return resourceSentryFilterRead(ctx, d, meta)
}

func resourceSentryFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*sentry.Client)

	org := d.Get("organization").(string)
	project := d.Get("project").(string)

	tflog.Debug(ctx, "Deleting Sentry filters browser extensions and legacy browser", map[string]interface{}{"org": org, "project": project})
	_, err := client.ProjectFilters.UpdateBrowserExtensions(ctx, org, project, false)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.ProjectFilters.UpdateLegacyBrowser(ctx, org, project, []string{})
	if err != nil {
		return diag.FromErr(err)
	}
	tflog.Debug(ctx, "Deleted Sentry filters browser extensions and legacy browser", map[string]interface{}{"org": org, "project": project})

	return nil
}
