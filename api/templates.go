package api

import "text/template"

// Templates as variables so we don't have to deal with paths.
// Could move this back out and add a config option though.
var (
	searchT = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:newznab="http://www.newznab.com/DTD/2010/feeds/attributes/">
  <channel>
    <atom:link href="{{.URL}}" rel="self" type="application/rss+xml" />
    <title>gonab API Search</title>
    <description>gonab Feed</description>
    <link>https://github.org/hobeone/gonab</link>
    <language>en-gb</language>
    <webMaster>{{.ContactEmail}}</webMaster>
    <category></category>
    <image>
      <url>{{.Image.URL}}</url>
      <title>{{.Image.Title}}</title>
      <link>{{.Image.Link}}</link>
      <description>Go Based NZB Harvester</description>
    </image>
    <newznab:response offset="{{.Offset}}" total="{{.Total}}" />
    {{ range .NZBs}}
    <item>
      <title>{{.Title}}</title>
      <guid isPermaLink="{{.PermaLink}}">{{.GUID}}</guid>
      <link>{{.Link}}</link>
      <comments>{{.Comments}}</comments>
      <pubDate>{{.Date}}</pubDate> 
      <category>{{.Category}}</category>  
      <description>{{.Title}}</description>
      <enclosure url="{{.Link}}" length="{{.Size}}" type="application/x-nzb" />
      <newznab:attr name="guid" value="{{.GUID}}" />
      <newznab:attr name="details" value="" />
      <newznab:attr name="category" value="5000" />
      <newznab:attr name="category" value="5040" />
      <newznab:attr name="size" value="{{.Size}}" />
    </item>
    {{- end }}
  </channel>
</rss>`
	searchResponseTemplate = template.Must(template.New("searchresponse").Parse(searchT))

	capsT = `<?xml version="1.0" encoding="UTF-8"?>
<caps>
  <server version="{{.Server.AppVersion}}" title="{{.Server.Title}}" strapline="{{.Server.Strapline}}" email="{{.Server.Email}}" url="{{.Server.URL}}" image="{{.Server.Image}}"/>
  <limits max="" default=""/>
  <retention days=""/>
  <registration available="{{.Registration.Available}}" open="{{.Registration.Open}}" />
  <searching>
	{{- range .Searching}}
    <{{.Name}} available="{{.Available}}" supportedParams="{{.SupportedParams}}"/>
	{{- end}}
  </searching>
  <categories>
    {{- range .Categories}}
    <category id="{{.ID}}" name="{{.Name}}">
      {{- range .SubCategories}}
      <subcat id="{{.ID}}" name="{{.Name}}"/>
      {{- end}}
    </category>
    {{- end}}
  </categories>
</caps>
</xml>`
	capsResponseTemplate = template.Must(template.New("capsresponse").Parse(capsT))
)
