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
 * @interface AddLocationResult
 */
export interface AddLocationResult {
    /**
     * 
     * @type {string}
     * @memberof AddLocationResult
     */
    uuid: string;
}

/**
 * Check if a given object implements the AddLocationResult interface.
 */
export function instanceOfAddLocationResult(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "uuid" in value;

    return isInstance;
}

export function AddLocationResultFromJSON(json: any): AddLocationResult {
    return AddLocationResultFromJSONTyped(json, false);
}

export function AddLocationResultFromJSONTyped(json: any, ignoreDiscriminator: boolean): AddLocationResult {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'uuid': json['uuid'],
    };
}

export function AddLocationResultToJSON(value?: AddLocationResult | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'uuid': value.uuid,
    };
}

