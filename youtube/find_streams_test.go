package youtube

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindStreams(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s := service{
			baseURL: "https://api.piped.projectsegfau.lt",
			http:    http.DefaultClient,
		}

		streams, err := s.FindStreams(context.Background(), "YAmLMohrus4")
		require.NoError(t, err)

		assert.Len(t, streams.AudioStreams, 6)
		assert.Len(t, streams.VideoStreams, 17)

		assert.Equal(t, "https://proxy.piped.projectsegfau.lt/videoplayback?bui=AWRWj2ThES0-NvPubGusiswqwHLALFBv_h_eJwBAO18kuR4AtYaiMUWAeQNuJ0QotnyCPGi01IPx_yRM&c=WEB&clen=3378083&cpn=q7JzVkVTYROK4MtD&dur=205.221&ei=jOIqZu-BA6jwi9oPzYaR4Ao&expire=1714108140&fvip=5&gir=yes&host=rr1---sn-4g5lznez.googlevideo.com&id=o-AKB2-bHTxXrJnQzQtLULrjrmhW5TVfYIMpLAm_X8vTnQ&initcwndbps=5913750&ip=2a0d%3A5940%3A99%3A3%3A709d%3A4c4b%3A1517%3Adf99&itag=251&keepalive=yes&lmt=1714074834384333&lsig=AHWaYeowRgIhAMxzGYnwcPIg2RbQIBTfAov0RwXxHB47rrj3RgJMtk3eAiEAyyQXlFZqswFCAatjUdsHBcu5kCg-gk6YTBJVTNI_3QY%3D&lsparams=mh%2Cmm%2Cmn%2Cms%2Cmv%2Cmvi%2Cpl%2Cinitcwndbps&mh=GH&mime=audio%2Fwebm&mm=31%2C26&mn=sn-4g5lznez%2Csn-25glene6&ms=au%2Conr&mt=1714086304&mv=m&mvi=1&n=vpO-PdvUC0c0Ag&ns=8DsdGAA3LddjGpZuAvpFRN0Q&pl=29&requiressl=yes&sefc=1&sig=AJfQdSswRQIhAPYOASVbUKXLYhVcyuiQ37djia1zNvMW_kiqgbQgLl_JAiA4AGTUboBivRBBaC9v5g4lUUz89kc3JWIUU3s7XRsfZA%3D%3D&source=youtube&sparams=expire%2Cei%2Cip%2Cid%2Citag%2Csource%2Crequiressl%2Cxpc%2Cbui%2Cspc%2Cvprv%2Csvpuc%2Cmime%2Cns%2Cgir%2Cclen%2Cdur%2Clmt&spc=UWF9f6v5XeyAdtyAk2RxWN5ZHbRjofjMCzCwI13rnM210vT5tJoDBK_je0rF&svpuc=1&txp=4532434&vprv=1&xpc=EgVo2aDSNQ%3D%3D", streams.AudioStreams[3].URL)
		assert.Equal(t, "129 kbps", streams.AudioStreams[3].Quality)
		assert.Equal(t, 3378083, streams.AudioStreams[3].ContentLength)
	})
}
