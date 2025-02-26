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
import type { CreateLocationRequestBodyConfig } from './CreateLocationRequestBodyConfig';
import {
    CreateLocationRequestBodyConfigFromJSON,
    CreateLocationRequestBodyConfigFromJSONTyped,
    CreateLocationRequestBodyConfigToJSON,
} from './CreateLocationRequestBodyConfig';

/**
 * 
 * @export
 * @interface CreateLocationRequestBody
 */
export interface CreateLocationRequestBody {
    /**
     * 
     * @type {CreateLocationRequestBodyConfig}
     * @memberof CreateLocationRequestBody
     */
    config?: CreateLocationRequestBodyConfig;
    /**
     * 
     * @type {string}
     * @memberof CreateLocationRequestBody
     */
    description?: string;
    /**
     * 
     * @type {string}
     * @memberof CreateLocationRequestBody
     */
    name: string;
    /**
     * 
     * @type {string}
     * @memberof CreateLocationRequestBody
     */
    purpose: CreateLocationRequestBodyPurposeEnum;
    /**
     * 
     * @type {string}
     * @memberof CreateLocationRequestBody
     */
    source: CreateLocationRequestBodySourceEnum;
}


/**
 * @export
 */
export const CreateLocationRequestBodyPurposeEnum = {
    Unspecified: 'unspecified',
    AipStore: 'aip_store'
} as const;
export type CreateLocationRequestBodyPurposeEnum = typeof CreateLocationRequestBodyPurposeEnum[keyof typeof CreateLocationRequestBodyPurposeEnum];

/**
 * @export
 */
export const CreateLocationRequestBodySourceEnum = {
    Unspecified: 'unspecified',
    Minio: 'minio',
    Sftp: 'sftp',
    Amss: 'amss'
} as const;
export type CreateLocationRequestBodySourceEnum = typeof CreateLocationRequestBodySourceEnum[keyof typeof CreateLocationRequestBodySourceEnum];


/**
 * Check if a given object implements the CreateLocationRequestBody interface.
 */
export function instanceOfCreateLocationRequestBody(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "name" in value;
    isInstance = isInstance && "purpose" in value;
    isInstance = isInstance && "source" in value;

    return isInstance;
}

export function CreateLocationRequestBodyFromJSON(json: any): CreateLocationRequestBody {
    return CreateLocationRequestBodyFromJSONTyped(json, false);
}

export function CreateLocationRequestBodyFromJSONTyped(json: any, ignoreDiscriminator: boolean): CreateLocationRequestBody {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'config': !exists(json, 'config') ? undefined : CreateLocationRequestBodyConfigFromJSON(json['config']),
        'description': !exists(json, 'description') ? undefined : json['description'],
        'name': json['name'],
        'purpose': json['purpose'],
        'source': json['source'],
    };
}

export function CreateLocationRequestBodyToJSON(value?: CreateLocationRequestBody | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'config': CreateLocationRequestBodyConfigToJSON(value.config),
        'description': value.description,
        'name': value.name,
        'purpose': value.purpose,
        'source': value.source,
    };
}

