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
import type { EnduroPackagePreservationAction } from './EnduroPackagePreservationAction';
import {
    EnduroPackagePreservationActionFromJSON,
    EnduroPackagePreservationActionFromJSONTyped,
    EnduroPackagePreservationActionToJSON,
} from './EnduroPackagePreservationAction';

/**
 * 
 * @export
 * @interface EnduroPackagePreservationActions
 */
export interface EnduroPackagePreservationActions {
    /**
     * 
     * @type {Array<EnduroPackagePreservationAction>}
     * @memberof EnduroPackagePreservationActions
     */
    actions?: Array<EnduroPackagePreservationAction>;
}

/**
 * Check if a given object implements the EnduroPackagePreservationActions interface.
 */
export function instanceOfEnduroPackagePreservationActions(value: object): boolean {
    let isInstance = true;

    return isInstance;
}

export function EnduroPackagePreservationActionsFromJSON(json: any): EnduroPackagePreservationActions {
    return EnduroPackagePreservationActionsFromJSONTyped(json, false);
}

export function EnduroPackagePreservationActionsFromJSONTyped(json: any, ignoreDiscriminator: boolean): EnduroPackagePreservationActions {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'actions': !exists(json, 'actions') ? undefined : ((json['actions'] as Array<any>).map(EnduroPackagePreservationActionFromJSON)),
    };
}

export function EnduroPackagePreservationActionsToJSON(value?: EnduroPackagePreservationActions | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'actions': value.actions === undefined ? undefined : ((value.actions as Array<any>).map(EnduroPackagePreservationActionToJSON)),
    };
}

