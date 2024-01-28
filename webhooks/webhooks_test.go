/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package webhooks

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Example_newEventListener() {
	_ = NewEventListener(
		WithNotificationErrorHandler(NoOpNotificationErrorHandler),
		WithAfterFunc(func(ctx context.Context, notification *Notification, err error) {}),
		WithBeforeFunc(func(ctx context.Context, notification *Notification) error {
			return nil
		}),
		WithGlobalNotificationHandler(
			func(ctx context.Context, writer http.ResponseWriter, notification *Notification) error {
				return nil
			}),
		WithHooks(&Hooks{
			OnOrderMessageHook:        nil,
			OnButtonMessageHook:       nil,
			OnLocationMessageHook:     nil,
			OnContactsMessageHook:     nil,
			OnMessageReactionHook:     nil,
			OnUnknownMessageHook:      nil,
			OnProductEnquiryHook:      nil,
			OnInteractiveMessageHook:  nil,
			OnMessageErrorsHook:       nil,
			OnTextMessageHook:         nil,
			OnReferralMessageHook:     nil,
			OnCustomerIDChangeHook:    nil,
			OnSystemMessageHook:       nil,
			OnMediaMessageHook:        nil,
			OnNotificationErrorHook:   nil,
			OnMessageStatusChangeHook: nil,
			OnMessageReceivedHook:     nil,
		}),
		WithSubscriptionVerifier(func(ctx context.Context, request *VerificationRequest) error {
			return fmt.Errorf("subscription verification failed")
		}),
		WithHandlerOptions(&HandlerOptions{
			BeforeFunc:        nil,
			AfterFunc:         nil,
			ValidateSignature: false,
			Secret:            "lilsecretofold",
		}),
		WithHooksErrorHandler(NoOpHooksErrorHandler),
	)

	// Output:
}

func TestParseMessageType(t *testing.T) {
	t.Parallel()
	type args struct {
		messageType string
	}

	tests := []struct {
		name string
		args args
		want MessageType
	}{
		{
			name: "tExt",
			args: args{
				messageType: "text",
			},
			want: TextMessageType,
		},
		{
			name: "imageX",
			args: args{
				messageType: "imageX",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseMessageType(tt.args.messageType)
			if got != tt.want {
				t.Errorf("ParseMessageType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// humanReadableNotification is a human-readable representation of a notification.
func humanReadableNotification(t *testing.T, notification *Notification) string {
	t.Helper()
	var buf bytes.Buffer
	if notification == nil {
		return "notification: <nil>"
	}

	buf.WriteString("notification: object: ")
	if notification.Object == "" {
		buf.WriteString("<nil>")
	} else {
		buf.WriteString(notification.Object)
	}
	buf.WriteString(", entries: ")
	for i, entry := range notification.Entry {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(humanReadableEntry(t, entry))
	}

	return buf.String()
}

func humanReadableEntry(t *testing.T, entry *Entry) string {
	t.Helper()
	var buf bytes.Buffer
	if entry == nil {
		return "entry: <nil>"
	}
	buf.WriteString("entry: id: ")
	if entry.ID == "" {
		buf.WriteString("<nil>")
	} else {
		buf.WriteString(entry.ID)
	}
	buf.WriteString(", changes: ")
	for i, change := range entry.Changes {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(humanReadableChange(t, change))
	}

	return buf.String()
}

func humanReadableChange(t *testing.T, change *Change) string {
	t.Helper()

	return fmt.Sprintf("change: field: %s, value: %+v", change.Field, change.Value)
}

func TestNotificationHandler_Options(t *testing.T) {
	t.Parallel()
	type fields struct {
		BeforeFunc        BeforeFunc
		AfterFunc         AfterFunc
		ValidateSignature bool
		Secret            string
		Hooks             *Hooks
		Body              []byte
	}

	testcases := []struct {
		name       string
		fields     fields
		wantStatus int
	}{
		{
			name: "no options",
			fields: fields{
				BeforeFunc:        nil,
				AfterFunc:         nil,
				ValidateSignature: false,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "with options",
			fields: fields{
				BeforeFunc: func(ctx context.Context, notification *Notification) error {
					// return NewFatalError(errors.New("fatal error"), "just being rude")
					return nil
				},
				AfterFunc: func(ctx context.Context, notification *Notification, err error) {
					t.Logf("notification: %+v", humanReadableNotification(t, notification))
				},
				ValidateSignature: false,
				Secret:            "demo",
				Hooks:             nil,
				Body:              []byte(`{"object":"whatsapp_business_account","entry":[{"id":"WHATSAPP_BUSINESS_ACCOUNT_ID","changes":[{"value":{"messaging_product":"whatsapp","metadata":{"display_phone_number":"PHONE_NUMBER","phone_number_id":"PHONE_NUMBER_ID"},"contacts":[{"profile":{"name":"NAME"},"wa_id":"WHATSAPP_ID"}],"messages":[{"from":"PHONE_NUMBER","id":"wamid.ID","timestamp":"TIMESTAMP","type":"image","image":{"caption":"CAPTION","mime_type":"image/jpeg","sha256":"IMAGE_HASH","id":"ID"}}]},"field":"messages"}]}]}`), //nolint:lll
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "order webhook",
			fields: fields{
				BeforeFunc: func(ctx context.Context, notification *Notification) error {
					// return NewFatalError(errors.New("fatal error"), "just being rude")
					return nil
				},
				AfterFunc: func(ctx context.Context, notification *Notification, err error) {
					t.Logf("notification: %+v", humanReadableNotification(t, notification))
				},
				ValidateSignature: false,
				Secret:            "demo",
				Hooks:             nil,
				Body:              []byte(`{"object":"whatsapp_business_account","entry":[{"id":"130363306827170","changes":[{"value":{"messaging_product":"whatsapp","metadata":{"display_phone_number":"6281388288202","phone_number_id":"175174709002390"},"contacts":[{"profile":{"name":"Ahmad Saekoni"},"wa_id":"6281272128270"}],"messages":[{"from":"6281272128270","id":"wamid.HBgNNjI4MTI3MjEyODI3MBUCABIYFjNFQjA0RDhBMjNCN0E0QzkyQjg0NEQA","timestamp":"1706460409","type":"order","order":{"catalog_id":"363547682948433","text":"","product_items":[{"product_retailer_id":"1710","quantity":1,"item_price":11000,"currency":"IDR"}]}}]},"field":"messages"}]}]}`), //nolint:lll
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "text webhook",
			fields: fields{
				BeforeFunc: func(ctx context.Context, notification *Notification) error {
					// return NewFatalError(errors.New("fatal error"), "just being rude")
					return nil
				},
				AfterFunc: func(ctx context.Context, notification *Notification, err error) {
					t.Logf("notification: %+v", humanReadableNotification(t, notification))
				},
				ValidateSignature: false,
				Secret:            "demo",
				Hooks:             nil,
				Body:              []byte(`{"object":"whatsapp_business_account","entry":[{"id":"144509515401993","changes":[{"value":{"messaging_product":"whatsapp","metadata":{"display_phone_number":"15550416043","phone_number_id":"121720824363144"},"contacts":[{"profile":{"name":"Ahmad Saekoni"},"wa_id":"6281272128270"}],"messages":[{"from":"6281272128270","id":"wamid.HBgNNjI4MTI3MjEyODI3MBUCABIYFjNFQjAzRjY1RTM0RkI3N0Q2QzE3MDEA","timestamp":"1706461964","text":{"body":"a"},"type":"text"}]},"field":"messages"}]}]}`), //nolint:lll
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "document webhook",
			fields: fields{
				BeforeFunc: func(ctx context.Context, notification *Notification) error {
					// return NewFatalError(errors.New("fatal error"), "just being rude")
					return nil
				},
				AfterFunc: func(ctx context.Context, notification *Notification, err error) {
					t.Logf("notification: %+v", humanReadableNotification(t, notification))
				},
				ValidateSignature: false,
				Secret:            "demo",
				Hooks:             nil,
				Body:              []byte(`{"object":"whatsapp_business_account","entry":[{"id":"144509515401993","changes":[{"value":{"messaging_product":"whatsapp","metadata":{"display_phone_number":"15550416043","phone_number_id":"121720824363144"},"contacts":[{"profile":{"name":"Ahmad Saekoni"},"wa_id":"6281272128270"}],"messages":[{"from":"6281272128270","id":"wamid.HBgNNjI4MTI3MjEyODI3MBUCABIYFjNFQjA0QjU3MjA0NDk4NkExRTc2NjQA","timestamp":"1706462002","type":"document","document":{"filename":"data_product (2).csv","mime_type":"text\/csv","sha256":"iIiDLz5Lp5qtydahqNdzk5z\/zXUpwJ68J\/b3Wkwhoos=","id":"1726124224580101"}}]},"field":"messages"}]}]}`), //nolint:lll
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "image webhook",
			fields: fields{
				BeforeFunc: func(ctx context.Context, notification *Notification) error {
					// return NewFatalError(errors.New("fatal error"), "just being rude")
					return nil
				},
				AfterFunc: func(ctx context.Context, notification *Notification, err error) {
					t.Logf("notification: %+v", humanReadableNotification(t, notification))
				},
				ValidateSignature: false,
				Secret:            "demo",
				Hooks:             nil,
				Body:              []byte(`{"object":"whatsapp_business_account","entry":[{"id":"144509515401993","changes":[{"value":{"messaging_product":"whatsapp","metadata":{"display_phone_number":"15550416043","phone_number_id":"121720824363144"},"contacts":[{"profile":{"name":"Ahmad Saekoni"},"wa_id":"6281272128270"}],"messages":[{"from":"6281272128270","id":"wamid.HBgNNjI4MTI3MjEyODI3MBUCABIYFjNFQjAzOUQ2NkZGMUJFMDAzNDJGNDMA","timestamp":"1706462045","type":"image","image":{"mime_type":"image\/jpeg","sha256":"hgAkC3q+WZRfX+5GNvi+DrwnSXq4zWveuUIELCQtPxE=","id":"923699576041120"}}]},"field":"messages"}]}]}`), //nolint:lll
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "contact webhook",
			fields: fields{
				BeforeFunc: func(ctx context.Context, notification *Notification) error {
					// return NewFatalError(errors.New("fatal error"), "just being rude")
					return nil
				},
				AfterFunc: func(ctx context.Context, notification *Notification, err error) {
					t.Logf("notification: %+v", humanReadableNotification(t, notification))
				},
				ValidateSignature: false,
				Secret:            "demo",
				Hooks:             nil,
				Body:              []byte(`{"object":"whatsapp_business_account","entry":[{"id":"144509515401993","changes":[{"value":{"messaging_product":"whatsapp","metadata":{"display_phone_number":"15550416043","phone_number_id":"121720824363144"},"contacts":[{"profile":{"name":"Ahmad Saekoni"},"wa_id":"6281272128270"}],"messages":[{"from":"6281272128270","id":"wamid.HBgNNjI4MTI3MjEyODI3MBUCABIYFjNFQjAzRjY1RTM0RkI3N0Q2QzE3MDEA","timestamp":"1706461964","text":{"body":"a"},"type":"text"}]},"field":"messages"}]}]}`), //nolint:lll
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range testcases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hooks := tt.fields.Hooks
			options := &HandlerOptions{
				BeforeFunc:        tt.fields.BeforeFunc,
				AfterFunc:         tt.fields.AfterFunc,
				ValidateSignature: tt.fields.ValidateSignature,
				Secret:            tt.fields.Secret,
			}
			h := NotificationHandler(hooks, NoOpNotificationErrorHandler, NoOpHooksErrorHandler, options)

			req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, "/webhook", bytes.NewReader(tt.fields.Body))
			if err != nil {
				t.Logf("error creating request: %v", err)
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}
		})
	}
}

func TestExtractSignatureFromHeader(t *testing.T) {
	t.Parallel()
	type args struct {
		header map[string][]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid signature",
			args: args{
				header: map[string][]string{
					SignatureHeaderKey: {"sha256=1234567890"},
				},
			},
			want:    "1234567890",
			wantErr: false,
		},
		{
			name: "invalid signature",
			args: args{
				header: map[string][]string{},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		name := tt.name
		args := tt.args
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, err := ExtractSignatureFromHeader(args.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractSignatureFromHeader() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if got != tt.want {
				t.Errorf("ExtractSignatureFromHeader() got = %v, want %v", got, tt.want)
			}
		})
	}
}
