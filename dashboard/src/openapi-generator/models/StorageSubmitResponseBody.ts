/* tslint:disable */
/* eslint-disable */
/**
 * Enduro API
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { exists, mapValues } from '../runtime';
/**
 * 
 * @export
 * @interface StorageSubmitResponseBody
 */
export interface StorageSubmitResponseBody {
    /**
     * 
     * @type {string}
     * @memberof StorageSubmitResponseBody
     */
    url: string;
}

export function StorageSubmitResponseBodyFromJSON(json: any): StorageSubmitResponseBody {
    return StorageSubmitResponseBodyFromJSONTyped(json, false);
}

export function StorageSubmitResponseBodyFromJSONTyped(json: any, ignoreDiscriminator: boolean): StorageSubmitResponseBody {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'url': json['url'],
    };
}

export function StorageSubmitResponseBodyToJSON(value?: StorageSubmitResponseBody | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'url': value.url,
    };
}
