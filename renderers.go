package main

import (
	"html/template"
	"os"
	"time"
)

func RenderIndex(layout Layout, templates *template.Template, out *os.File) error {
	layout.Title = "GCP Quickstart"
	layout.Source = Source{
		IsFreeForm: true,
		RenderedHTML: template.HTML(`
    <div>
      <amp-img
        width="917"
        height="318px"
        alt="cloud-logo"
        layout="responsive"
        src="/img/gcp.png"></app-img>
    </div>

    <h3>Get started with GCP</h3>
    <p>
      This is an open quickstart guide to Google Cloud Platform.
      All the learnings contained on this webpage are 100% open and liscensed
      under the gplv3. If you'd like to fix any typos or suggest an edit,
      you can file an issue on <a href="example.com">github</a>, or
      comment on the relavent google doc. At the top left of every quickstart
      you'll see a button linking to a google doc. Simply add a comment on
      the relavent content in the google doc, and we'll be sure to address
      your issue.
    </p>

    <h3>Why an Open Guide?</h3>
    <p>
      Closed documentation fails to iterate and keep itself up to date. The
      hope is that by being open we can foster a relationship with the GCP
      community to make the best documents possible, through iteration. The
      reason for choosing the GPLv3 is that we should all benefit by any improvements
      made to this documentation.
    </p>

    <h3>How is this different than cloud.google.com?</h3>
    <p>
      All documentation on this page serves to act as introduction material to
      Google Cloud Services. It aims to get you from nothing to having basic
      services deployed. All the material assumes that the end user has
      a basic understanding of the command line, and generic linux knowledge.
    </p>

    <h3>How can I Contribute</h3>
    <p>
      If you'd like to add content or become a major contributor simply post an
      issue to the <a href="example.com">github</a> repo. Be sure to include
      your Google Cloud subject
      matter expertise.
    </p>
    `),
	}

	layout.Social.Headline = "An Open Quickstart Guide to Google Cloud Platform"
	layout.Social.DatePublished = time.Now()
	layout.Social.Image = append(layout.Social.Image, "http://www.averesystems.com/cmsFiles/relatedImages/logo_lockup_cloud_platform_icon_vertical.png")

	return templates.ExecuteTemplate(out, "page", layout)
}
