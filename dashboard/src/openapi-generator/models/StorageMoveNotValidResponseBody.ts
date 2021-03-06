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
 * move_not_valid_response_body result type (default view)
 * @export
 * @interface StorageMoveNotValidResponseBody
 */
export interface StorageMoveNotValidResponseBody {
    /**
     * Is the error a server-side fault?
     * @type {boolean}
     * @memberof StorageMoveNotValidResponseBody
     */
    fault: boolean;
    /**
     * ID is a unique identifier for this particular occurrence of the problem.
     * @type {string}
     * @memberof StorageMoveNotValidResponseBody
     */
    id: string;
    /**
     * Message is a human-readable explanation specific to this occurrence of the problem.
     * @type {string}
     * @memberof StorageMoveNotValidResponseBody
     */
    message: string;
    /**
     * Name is the name of this class of errors.
     * @type {string}
     * @memberof StorageMoveNotValidResponseBody
     */
    name: string;
    /**
     * Is the error temporary?
     * @type {boolean}
     * @memberof StorageMoveNotValidResponseBody
     */
    temporary: boolean;
    /**
     * Is the error a timeout?
     * @type {boolean}
     * @memberof StorageMoveNotValidResponseBody
     */
    timeout: boolean;
}

export function StorageMoveNotValidResponseBodyFromJSON(json: any): StorageMoveNotValidResponseBody {
    return StorageMoveNotValidResponseBodyFromJSONTyped(json, false);
}

export function StorageMoveNotValidResponseBodyFromJSONTyped(json: any, ignoreDiscriminator: boolean): StorageMoveNotValidResponseBody {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'fault': json['fault'],
        'id': json['id'],
        'message': json['message'],
        'name': json['name'],
        'temporary': json['temporary'],
        'timeout': json['timeout'],
    };
}

export function StorageMoveNotValidResponseBodyToJSON(value?: StorageMoveNotValidResponseBody | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'fault': value.fault,
        'id': value.id,
        'message': value.message,
        'name': value.name,
        'temporary': value.temporary,
        'timeout': value.timeout,
    };
}

