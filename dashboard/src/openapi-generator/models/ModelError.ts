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
 * @interface ModelError
 */
export interface ModelError {
    /**
     * Is the error a server-side fault?
     * @type {boolean}
     * @memberof ModelError
     */
    fault: boolean;
    /**
     * ID is a unique identifier for this particular occurrence of the problem.
     * @type {string}
     * @memberof ModelError
     */
    id: string;
    /**
     * Message is a human-readable explanation specific to this occurrence of the problem.
     * @type {string}
     * @memberof ModelError
     */
    message: string;
    /**
     * Name is the name of this class of errors.
     * @type {string}
     * @memberof ModelError
     */
    name: string;
    /**
     * Is the error temporary?
     * @type {boolean}
     * @memberof ModelError
     */
    temporary: boolean;
    /**
     * Is the error a timeout?
     * @type {boolean}
     * @memberof ModelError
     */
    timeout: boolean;
}

/**
 * Check if a given object implements the ModelError interface.
 */
export function instanceOfModelError(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "fault" in value;
    isInstance = isInstance && "id" in value;
    isInstance = isInstance && "message" in value;
    isInstance = isInstance && "name" in value;
    isInstance = isInstance && "temporary" in value;
    isInstance = isInstance && "timeout" in value;

    return isInstance;
}

export function ModelErrorFromJSON(json: any): ModelError {
    return ModelErrorFromJSONTyped(json, false);
}

export function ModelErrorFromJSONTyped(json: any, ignoreDiscriminator: boolean): ModelError {
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

export function ModelErrorToJSON(value?: ModelError | null): any {
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

