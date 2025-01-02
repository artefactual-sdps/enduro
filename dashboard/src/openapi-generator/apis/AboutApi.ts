/* tslint:disable */
/* eslint-disable */
/**
 * Enduro API
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 0.0.1
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */


import * as runtime from '../runtime';
import type {
  EnduroAbout,
} from '../models/index';
import {
    EnduroAboutFromJSON,
    EnduroAboutToJSON,
} from '../models/index';

/**
 * AboutApi - interface
 * 
 * @export
 * @interface AboutApiInterface
 */
export interface AboutApiInterface {
    /**
     * Get information about the system
     * @summary about about
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     * @memberof AboutApiInterface
     */
    aboutAboutRaw(initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<EnduroAbout>>;

    /**
     * Get information about the system
     * about about
     */
    aboutAbout(initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<EnduroAbout>;

}

/**
 * 
 */
export class AboutApi extends runtime.BaseAPI implements AboutApiInterface {

    /**
     * Get information about the system
     * about about
     */
    async aboutAboutRaw(initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<EnduroAbout>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        if (this.configuration && this.configuration.accessToken) {
            const token = this.configuration.accessToken;
            const tokenString = await token("jwt_header_Authorization", []);

            if (tokenString) {
                headerParameters["Authorization"] = `Bearer ${tokenString}`;
            }
        }
        const response = await this.request({
            path: `/about`,
            method: 'GET',
            headers: headerParameters,
            query: queryParameters,
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => EnduroAboutFromJSON(jsonValue));
    }

    /**
     * Get information about the system
     * about about
     */
    async aboutAbout(initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<EnduroAbout> {
        const response = await this.aboutAboutRaw(initOverrides);
        return await response.value();
    }

}