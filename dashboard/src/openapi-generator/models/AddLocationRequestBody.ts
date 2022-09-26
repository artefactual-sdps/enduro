/* tslint:disable */
/* eslint-disable */
/**
 * Enduro API
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 1.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { exists, mapValues } from '../runtime';
import type { AddLocationRequestBodyConfig } from './AddLocationRequestBodyConfig';
import {
    AddLocationRequestBodyConfigFromJSON,
    AddLocationRequestBodyConfigFromJSONTyped,
    AddLocationRequestBodyConfigToJSON,
} from './AddLocationRequestBodyConfig';

/**
 * 
 * @export
 * @interface AddLocationRequestBody
 */
export interface AddLocationRequestBody {
    /**
     * 
     * @type {AddLocationRequestBodyConfig}
     * @memberof AddLocationRequestBody
     */
    config?: AddLocationRequestBodyConfig;
    /**
     * 
     * @type {string}
     * @memberof AddLocationRequestBody
     */
    description?: string;
    /**
     * 
     * @type {string}
     * @memberof AddLocationRequestBody
     */
    name: string;
    /**
     * 
     * @type {string}
     * @memberof AddLocationRequestBody
     */
    purpose: AddLocationRequestBodyPurposeEnum;
    /**
     * 
     * @type {string}
     * @memberof AddLocationRequestBody
     */
    source: AddLocationRequestBodySourceEnum;
}


/**
 * @export
 */
export const AddLocationRequestBodyPurposeEnum = {
    Unspecified: 'unspecified',
    AipStore: 'aip_store'
} as const;
export type AddLocationRequestBodyPurposeEnum = typeof AddLocationRequestBodyPurposeEnum[keyof typeof AddLocationRequestBodyPurposeEnum];

/**
 * @export
 */
export const AddLocationRequestBodySourceEnum = {
    Unspecified: 'unspecified',
    Minio: 'minio'
} as const;
export type AddLocationRequestBodySourceEnum = typeof AddLocationRequestBodySourceEnum[keyof typeof AddLocationRequestBodySourceEnum];


/**
 * Check if a given object implements the AddLocationRequestBody interface.
 */
export function instanceOfAddLocationRequestBody(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "name" in value;
    isInstance = isInstance && "purpose" in value;
    isInstance = isInstance && "source" in value;

    return isInstance;
}

export function AddLocationRequestBodyFromJSON(json: any): AddLocationRequestBody {
    return AddLocationRequestBodyFromJSONTyped(json, false);
}

export function AddLocationRequestBodyFromJSONTyped(json: any, ignoreDiscriminator: boolean): AddLocationRequestBody {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'config': !exists(json, 'config') ? undefined : AddLocationRequestBodyConfigFromJSON(json['config']),
        'description': !exists(json, 'description') ? undefined : json['description'],
        'name': json['name'],
        'purpose': json['purpose'],
        'source': json['source'],
    };
}

export function AddLocationRequestBodyToJSON(value?: AddLocationRequestBody | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'config': AddLocationRequestBodyConfigToJSON(value.config),
        'description': value.description,
        'name': value.name,
        'purpose': value.purpose,
        'source': value.source,
    };
}
