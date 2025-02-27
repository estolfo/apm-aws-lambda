// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.uber.org/zap"
)

func loadAWSOptions(ctx context.Context, cfg aws.Config, logger *zap.SugaredLogger) (string, string, error) {
	manager := secretsmanager.NewFromConfig(cfg)

	apmServerApiKey := os.Getenv("ELASTIC_APM_API_KEY")
	if apmServerApiKeySMSecretId, ok := os.LookupEnv("ELASTIC_APM_SECRETS_MANAGER_API_KEY_ID"); ok {
		result, err := loadSecret(ctx, manager, apmServerApiKeySMSecretId)
		if err != nil {
			return "", "", fmt.Errorf("failed loading APM Server ApiKey from Secrets Manager: %w", err)
		}

		logger.Infof("Using the APM API key retrieved from Secrets Manager.")
		apmServerApiKey = result
	}

	apmServerSecretToken := os.Getenv("ELASTIC_APM_SECRET_TOKEN")
	if apmServerSecretTokenSMSecretId, ok := os.LookupEnv("ELASTIC_APM_SECRETS_MANAGER_SECRET_TOKEN_ID"); ok {
		result, err := loadSecret(ctx, manager, apmServerSecretTokenSMSecretId)
		if err != nil {
			return "", "", fmt.Errorf("failed loading APM Server Secret Token from Secrets Manager: %w", err)
		}

		logger.Infof("Using the APM secret token retrieved from Secrets Manager.")
		apmServerSecretToken = result
	}

	return apmServerApiKey, apmServerSecretToken, nil
}

func loadSecret(ctx context.Context, manager *secretsmanager.Client, secretID string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     ptrFromString(secretID),
		VersionStage: ptrFromString("AWSCURRENT"),
	}

	result, err := manager.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve sercet value: %w", err)
	}

	if result.SecretString != nil {
		return *result.SecretString, nil
	}

	decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
	if _, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary); err != nil {
		return "", fmt.Errorf("failed to decode base64 encoded secret: %w", err)
	}

	return string(decodedBinarySecretBytes), nil
}

func ptrFromString(v string) *string {
	return &v
}
