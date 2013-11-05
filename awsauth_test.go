package awsauth

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestIntegration(t *testing.T) {
	Convey("Given real credentials from environment variables", t, func() {
		Convey("A request to IAM should succeed", nil)

		Convey("A request to S3 should succeed", nil)

		Convey("A request to EC2 should succeed", func() {
			req := newRequest("GET", "https://ec2.amazonaws.com", url.Values{
				"Action": []string{"DescribeInstances"},
			})
			resp := sign2AndDo(req)

			if !envCredentialsSet() {
				SkipSo(resp.StatusCode, ShouldEqual, http.StatusOK)
			} else {
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
			}
		})

		Convey("A request to SQS should succeed", func() {
			req := newRequest("POST", "https://sqs.us-west-2.amazonaws.com", url.Values{
				"Action": []string{"ListQueues"},
			})
			resp := sign4AndDo(req)

			if !envCredentialsSet() {
				SkipSo(resp.StatusCode, ShouldEqual, http.StatusOK)
			} else {
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
			}
		})
	})
}

func TestSign(t *testing.T) {
	Convey("Requests to services using Version 2 should be signed accordingly", t, func() {
		reqs := []*http.Request{
			newRequest("GET", "https://ec2.amazonaws.com", url.Values{}),
			newRequest("GET", "https://elasticache.amazonaws.com/", url.Values{}),
		}
		for _, req := range reqs {
			signedReq := Sign(req)
			So(signedReq.URL.Query().Get("SignatureVersion"), ShouldEqual, "2")
		}
	})

	Convey("Requests to services using Version 4 should be signed accordingly", t, func() {
		reqs := []*http.Request{
			newRequest("POST", "https://sqs.amazonaws.com/", url.Values{}),
			newRequest("GET", "https://iam.amazonaws.com", url.Values{}),
		}
		for _, req := range reqs {
			signedReq := Sign(req)
			So(signedReq.Header.Get("Authorization"), ShouldContainSubstring, ", Signature=")
		}
	})

	SkipConvey("Requests to S3 should be signed accordingly", t, func() {
		req := newRequest("POST", "https://s3.amazonaws.com", url.Values{})
		signedReq := Sign(req)
		So(signedReq.Header.Get("Authorization"), ShouldContainSubstring, "AWS ")
	})
}

func envCredentialsSet() bool {
	return os.Getenv(envAccessKeyID) != "" && os.Getenv(envSecretAccessKey) != ""
}

func newRequest(method string, url string, v url.Values) *http.Request {
	req, _ := http.NewRequest(method, url, strings.NewReader(v.Encode()))
	return req
}

func sign2AndDo(req *http.Request) *http.Response {
	Sign2(req)
	resp, _ := client.Do(req)
	return resp
}

func sign4AndDo(req *http.Request) *http.Response {
	Sign4(req)
	resp, _ := client.Do(req)
	return resp
}

var client = &http.Client{}
