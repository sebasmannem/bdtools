#!/bin/bash

function debug () {
    if [ "${DEBUG}" != "" ]
    then
        echo $@ 1>&2
    fi
}

function getQuotaOplosgroepen() {

    local QUOTAOPLOSGROEPEN=$(oc get clusterresourcequota -o jsonpath='{.items[*].spec.selector.labels.matchLabels.dcs\.itsmoplosgroep}')

    echo $QUOTAOPLOSGROEPEN
}

function getNamespaceOplosgroepen() {

    local NAMESPACEGROEPEN=$(oc get namespaces -o jsonpath='{.items[*].metadata.labels.dcs\.itsmoplosgroep}')

    echo $NAMESPACEGROEPEN
}

function compareOplosgroepen() {
    local TEMPQUOTAOPLOSGROEPENLIJST=$(getQuotaOplosgroepen)
    IFS=' ' read -r -a QUOTAOPLOSGROEPENLIJST <<< "${TEMPQUOTAOPLOSGROEPENLIJST}"

    local TEMPNAMESPACEOPLOSGROEPENLIJST=$(getNamespaceOplosgroepen)
    IFS=' ' read -r -a NAMESPACEOPLOSGROEPENLIJST <<< "${TEMPNAMESPACEOPLOSGROEPENLIJST}"
    DEDUPED_NAMESPACEOPLOSGROEPENLIJST=($(echo "${NAMESPACEOPLOSGROEPENLIJST[@]}" | tr ' ' '\n' | sort -u | tr '\n' ' '))

    PATCH_ARRAY=()
    for NAMESPACEOPLOSGROEP in "${DEDUPED_NAMESPACEOPLOSGROEPENLIJST[@]}"
    do
        OPLOSGROEP="false"
        for QUOTAOPLOSGROEP in "${QUOTAOPLOSGROEPENLIJST[@]}"
        do
            if [ "${NAMESPACEOPLOSGROEP}" == "${QUOTAOPLOSGROEP}" ];
            then
                OPLOSGROEP="true"
                echo "Namespace Oplosgroep: ${QUOTAOPLOSGROEP} exists as CRQ"
            fi
        done
        if [ "${OPLOSGROEP}" != "true" ];
        then
            echo "Oplosgroep: ${NAMESPACEOPLOSGROEP} does not exist in our CRQ's"
            MISSINGBLOCKRESOURCESLIST+=("${NAMESPACEOPLOSGROEP}")
        fi
    done
}

function patchOplosgroepen () {

    local TEMPCURRENTBLOCKRESOURCESLIST=$(oc get clusterresourcequota blockresources -o jsonpath='{.spec.selector.labels.matchExpressions[*].values[*]}')
    IFS=' ' read -r -a CURRENTBLOCKRESOURCESLIST <<< "${TEMPCURRENTBLOCKRESOURCESLIST}"

    # Call compareOplosgroepen function
    compareOplosgroepen

    for MISSINGBLOCKRESOURCE in "${MISSINGBLOCKRESOURCESLIST[@]}"
    do
        BLOCKRESOURCE="false"
        for CURRENTBLOCKRESOURCE in "${CURRENTBLOCKRESOURCESLIST[@]}"
        do
            if [ "${MISSINGBLOCKRESOURCE}" == "${CURRENTBLOCKRESOURCE}" ];
            then
                BLOCKRESOURCE="true"
                echo "Blocked Resource ${MISSINGBLOCKRESOURCE} already exists!"
            fi
        done
        if [ "${BLOCKRESOURCE}" != "true" ];
        then
            #TO DO For loop below example with != true MISSINGBLOCKRESOURCE's
            oc patch clusterresourcequota blockresources --type=json -p='[{"op": "add", "path": "/spec/selector/labels/matchExpressions/0/values/-", "value": "4"}]'
            echo "BlockedResource ${MISSINGBLOCKRESOURCE} patched"
        fi
    done
}

#Output manipulation (To Do)
GREEN='\033[0;32m'
ORANGE='\033[0;33m'
LRED='\033[1;31m'
NC='\033[0m'

#Initialize
patchOplosgroepen

