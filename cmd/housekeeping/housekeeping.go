// Copyright © 2019 Ispirata Srl
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package housekeeping

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/url"
	"path"
	"time"
)

// housekeepingCmd represents the housekeeping command
var HousekeepingCmd = &cobra.Command{
	Use:               "housekeeping",
	Short:             "Interact with Housekeeping API",
	Long:              `Interact with Housekeeping API.`,
	PersistentPreRunE: housekeepingPersistentPreRunE,
}

var housekeepingJwt string
var housekeepingUrl string

func init() {
	HousekeepingCmd.PersistentFlags().StringP("housekeeping-key", "k", "",
		"Path to housekeeping private key to generate JWT for authentication")
	HousekeepingCmd.MarkPersistentFlagFilename("housekeeping-key")
	viper.BindPFlag("housekeeping.key", HousekeepingCmd.PersistentFlags().Lookup("housekeeping-key"))
	HousekeepingCmd.PersistentFlags().String("housekeeping-url", "",
		"Housekeeping API base URL. Defaults to <astarte-url>/housekeeping.")
	viper.BindPFlag("housekeeping.url", HousekeepingCmd.PersistentFlags().Lookup("housekeeping-url"))
}

func housekeepingPersistentPreRunE(cmd *cobra.Command, args []string) error {
	housekeepingUrlOverride := viper.GetString("housekeeping.url")
	astarteUrl := viper.GetString("url")
	if housekeepingUrlOverride != "" {
		// Use explicit housekeeping-url
		housekeepingUrl = housekeepingUrlOverride
	} else if astarteUrl != "" {
		url, _ := url.Parse(astarteUrl)
		url.Path = path.Join(url.Path, "housekeeping")
		housekeepingUrl = url.String()
	} else {
		return errors.New("Either astarte-url or housekeeping-url have to be specified")
	}

	housekeepingKey := viper.GetString("housekeeping.key")
	if housekeepingKey == "" {
		return errors.New("housekeeping-key is required")
	}

	var err error
	housekeepingJwt, err = generateHousekeepingJWT(housekeepingKey)
	if err != nil {
		return err
	}

	return nil
}

func generateHousekeepingJWT(privateKey string) (jwtString string, err error) {
	keyPEM, err := ioutil.ReadFile(privateKey)
	if err != nil {
		return "", err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyPEM)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC().Unix()
	// 5 minutes expiry
	expiry := now + 300
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"a_ha": []string{"^.*$::^.*$"},
		"iat":  now,
		"exp":  expiry,
	})

	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
