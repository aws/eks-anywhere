// Code generated by smithy-go-codegen DO NOT EDIT.

package snowballdevice

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/eks-anywhere/pkg/aws/sdk/v2/service/snowballdevice/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

func (c *Client) DescribeDirectNetworkInterfaces(ctx context.Context, params *DescribeDirectNetworkInterfacesInput, optFns ...func(*Options)) (*DescribeDirectNetworkInterfacesOutput, error) {
	if params == nil {
		params = &DescribeDirectNetworkInterfacesInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "DescribeDirectNetworkInterfaces", params, optFns, c.addOperationDescribeDirectNetworkInterfacesMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*DescribeDirectNetworkInterfacesOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type DescribeDirectNetworkInterfacesInput struct {
	NextToken *string

	noSmithyDocumentSerde
}

type DescribeDirectNetworkInterfacesOutput struct {

	// This member is required.
	DirectNetworkInterfaces []types.DirectNetworkInterface

	NextToken *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationDescribeDirectNetworkInterfacesMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpDescribeDirectNetworkInterfaces{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpDescribeDirectNetworkInterfaces{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opDescribeDirectNetworkInterfaces(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opDescribeDirectNetworkInterfaces(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "snowballdevice",
		OperationName: "DescribeDirectNetworkInterfaces",
	}
}
