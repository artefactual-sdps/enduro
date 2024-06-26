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

import { exists, mapValues } from '../runtime';
/**
 * 
 * @export
 * @interface ConfirmRequestBody
 */
export interface ConfirmRequestBody {
    /**
     * Identifier of storage location
     * @type {string}
     * @memberof ConfirmRequestBody
     */
    locationId: string;
}

/**
 * Check if a given object implements the ConfirmRequestBody interface.
 */
export function instanceOfConfirmRequestBody(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "locationId" in value;

    return isInstance;
}

export function ConfirmRequestBodyFromJSON(json: any): ConfirmRequestBody {
    return ConfirmRequestBodyFromJSONTyped(json, false);
}

export function ConfirmRequestBodyFromJSONTyped(json: any, ignoreDiscriminator: boolean): ConfirmRequestBody {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'locationId': json['location_id'],
    };
}

export function ConfirmRequestBodyToJSON(value?: ConfirmRequestBody | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'location_id': value.locationId,
    };
}

