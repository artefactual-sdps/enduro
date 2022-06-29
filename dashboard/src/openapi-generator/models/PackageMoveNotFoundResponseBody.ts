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
 * Package not found
 * @export
 * @interface PackageMoveNotFoundResponseBody
 */
export interface PackageMoveNotFoundResponseBody {
    /**
     * Identifier of missing package
     * @type {number}
     * @memberof PackageMoveNotFoundResponseBody
     */
    id: number;
    /**
     * Message of error
     * @type {string}
     * @memberof PackageMoveNotFoundResponseBody
     */
    message: string;
}

export function PackageMoveNotFoundResponseBodyFromJSON(json: any): PackageMoveNotFoundResponseBody {
    return PackageMoveNotFoundResponseBodyFromJSONTyped(json, false);
}

export function PackageMoveNotFoundResponseBodyFromJSONTyped(json: any, ignoreDiscriminator: boolean): PackageMoveNotFoundResponseBody {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'id': json['id'],
        'message': json['message'],
    };
}

export function PackageMoveNotFoundResponseBodyToJSON(value?: PackageMoveNotFoundResponseBody | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'id': value.id,
        'message': value.message,
    };
}

