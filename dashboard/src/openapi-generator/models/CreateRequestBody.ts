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
 * @interface CreateRequestBody
 */
export interface CreateRequestBody {
    /**
     * Identifier of AIP
     * @type {string}
     * @memberof CreateRequestBody
     */
    aipId: string;
    /**
     * Identifier of the package's storage location
     * @type {string}
     * @memberof CreateRequestBody
     */
    locationId?: string;
    /**
     * Name of the package
     * @type {string}
     * @memberof CreateRequestBody
     */
    name: string;
    /**
     * ObjectKey of AIP
     * @type {string}
     * @memberof CreateRequestBody
     */
    objectKey: string;
    /**
     * Status of the package
     * @type {string}
     * @memberof CreateRequestBody
     */
    status?: CreateRequestBodyStatusEnum;
}


/**
 * @export
 */
export const CreateRequestBodyStatusEnum = {
    Unspecified: 'unspecified',
    InReview: 'in_review',
    Rejected: 'rejected',
    Stored: 'stored',
    Moving: 'moving'
} as const;
export type CreateRequestBodyStatusEnum = typeof CreateRequestBodyStatusEnum[keyof typeof CreateRequestBodyStatusEnum];


/**
 * Check if a given object implements the CreateRequestBody interface.
 */
export function instanceOfCreateRequestBody(value: object): boolean {
    let isInstance = true;
    isInstance = isInstance && "aipId" in value;
    isInstance = isInstance && "name" in value;
    isInstance = isInstance && "objectKey" in value;

    return isInstance;
}

export function CreateRequestBodyFromJSON(json: any): CreateRequestBody {
    return CreateRequestBodyFromJSONTyped(json, false);
}

export function CreateRequestBodyFromJSONTyped(json: any, ignoreDiscriminator: boolean): CreateRequestBody {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'aipId': json['aip_id'],
        'locationId': !exists(json, 'location_id') ? undefined : json['location_id'],
        'name': json['name'],
        'objectKey': json['object_key'],
        'status': !exists(json, 'status') ? undefined : json['status'],
    };
}

export function CreateRequestBodyToJSON(value?: CreateRequestBody | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'aip_id': value.aipId,
        'location_id': value.locationId,
        'name': value.name,
        'object_key': value.objectKey,
        'status': value.status,
    };
}

