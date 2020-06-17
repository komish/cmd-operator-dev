#!/usr/bin/env bash
# Quick and dirty script to check for the existence of things
# created by creating a CertManagerDeployment CustomResource

set -u

INSTALLNAMESPACE="cert-manager"
expected_sas=3
expected_roles=3
expected_rbs=3
expected_clusterroles=9
expected_clrbs=7
expected_deploys=3

# Check and Cross
CHK="\xE2\x9C\x94"
CRS="\xE2\x9C\x98"


## ========================================================================================
## (check) pre-requisite tools are installed on the system.
prereqs='oc grep egrep test wc'

for cmd in ${prereqs}; do 
  which "${cmd}" &>/dev/null || { echo "Missing command: ${cmd}"; exit 1 ;}
done
echo "Pre-requisites are satisfied."

## ========================================================================================
## (check) The current OpenShift context is valid.
resp=$(oc status 2>&1)
test $? -ne 0 && \
    { echo "The 'oc status' command did not return with a successful error code. Message was:";
      echo -e "${resp}"; 
      exit 1 ;}
echo "-- Access to openshift cluster confirmed with context: $(oc config current-context)"

## ========================================================================================
## (check) The context is in the expected namespace
if [[ "$(oc project -q)" != "${INSTALLNAMESPACE}" ]]; then
    echo "-- Current context namespace is incorrect. Expecting ${INSTALLNAMESPACE}.";
    exit 1
fi

## ========================================================================================
## (check) We have the expected number of service accounts.
## These service accounts are not configurable so we check for exactly the cert-manager name
actual_sas=$(oc get sa | grep cert-manager | wc -l)
test ${expected_sas} -ne ${actual_sas} && \
    { echo -e "${CRS} FAIL (Service Accounts): Expected ${expected_sas}, Found ${actual_sas}";}
echo -e "${CHK} PASS (Service Accounts)"

## (check) We have the expected number of roles.
actual_roles=$(oc get roles | egrep "cert-manager-(cainjector|controller|webhook)" | wc -l)
if [[ "${expected_roles}" -ne "${actual_roles}" ]]; then
    echo -e "${CRS} FAIL (Roles): Expected ${expected_roles}, Found ${actual_roles}"
else
    echo -e "${CHK} PASS (Roles)"
fi

## ========================================================================================
## (check) We have the expected number of rolebindings.
## rolebindings rely on the CR name so we check for the static pieces that are appended
actual_rbs=$(oc get rolebindings | egrep "cert-manager-" | wc -l)
if [[ ${expected_rbs} -ne ${actual_rbs} ]]; then 
    echo -e "${CRS} FAIL (RoleBindings): Expected ${expected_rbs}, Found $(echo ${actual_rbs} | sed -e 's/^[ \t]*//')"
else
    echo -e "${CHK} PASS (RoleBindings)"
fi

## ========================================================================================
## (check) Rolebindings that exist tie an existing role to an existing serviceaccount
rbs=$(oc get rolebindings -o name| egrep "cert-manager-")
for rb in ${rbs}; do 
    echo "-- Checking Roles and Subjects for RoleBinding: ${rb}"
    # do the roles exist?
    rbrole=$(oc get "${rb}" -o go-template='{{ .roleRef.name }}')
    if oc get role "${rbrole}" &>/dev/null; then
        echo -e "${CHK}  |-- PASS (RoleBindings:RoleRefCheck) Role: ${rbrole}"
    else
        echo -e "${CRS}  |-- FAIL (RoleBindings:RoleCheck) Role ${rbrole} referenced in rolebinding ${rb} was not found in cluster";
    fi

    # do the subjects exist
    rbsubs=$(oc get "${rb}" -o go-template='{{range .subjects}}{{ .name }}{{" "}}{{end}}')
    for rbsub in ${rbsubs}; do
        if ! oc get sa "${rbsub}" &>/dev/null; then
            echo -e "${CRS}  |-- FAIL (RoleBindings:SubjectCheck) Role ${rbrole} referenced in rolebinding ${rb} was not found in cluster";
        else 
            echo -e "${CHK}  |-- PASS (RoleBindings:SubjectCheck) Subject: ${rbsub}"
        fi
    done
done


## ========================================================================================
## (check) We have the expected number of cluster roles
actual_clusterroles=$(oc get clusterroles | egrep "cert-manager" | wc -l)
if [[ "${expected_clusterroles}" -ne "${actual_clusterroles}" ]]; then
    echo -e "${CRS} FAIL (ClusterRoles): Expected ${expected_clusterroles}, Found ${actual_clusterroles}"
else
    echo -e "${CHK} PASS (ClusterRoles)"
fi

## ========================================================================================
## (check) We have the expected number of clusterrolebindings
actual_clrbs=$(oc get clusterrolebindings | egrep "cert-manager-" | wc -l)
if [[ ${expected_clrbs} -ne ${actual_clrbs} ]]; then 
    echo -e "${CRS} FAIL (ClusterRoleBindings): Expected ${expected_clrbs}, Found $(echo ${actual_clrbs} | sed -e 's/^[ \t]*//')"
else
    echo -e "${CHK} PASS (ClusterRoleBindings)"
fi

## ========================================================================================
## (check) ClusterRoleBindings that exist tie an existing cluster role to an existing serviceaccount
clrbs=$(oc get clusterrolebindings -o name| egrep "cert-manager-")
for clrb in ${clrbs}; do 
    echo "-- Checking ClusterRoles and Subjects for ClusterRoleBinding: ${clrb}"
    # do the roles exist?
    clrbrole=$(oc get "${clrb}" -o go-template='{{ .roleRef.name }}')
    if oc get clusterrole "${clrbrole}" &>/dev/null; then
        echo -e "${CHK}  |-- PASS (ClusterRoleBindings:RoleRefCheck) Role: ${clrbrole}"
    else
        echo -e "${CRS}  |-- FAIL (ClusterRoleBindings:RoleCheck) Role ${clrbrole} referenced in rolebinding ${clrb} was not found in cluster";
    fi

    # do the subjects exist
    clrbsubs=$(oc get "${clrb}" -o go-template='{{range .subjects}}{{ .name }}{{" "}}{{end}}')
    for clrbsub in ${clrbsubs}; do
        if ! oc get sa "${clrbsub}" &>/dev/null; then
            echo -e "${CRS}  |-- FAIL (ClusterRoleBindings:SubjectCheck) ClusterRole ${clrbrole} referenced in ClusterRoleBinding ${clrb} was not found in cluster";
        else 
            echo -e "${CHK}  |-- PASS (ClusterRoleBindings:SubjectCheck) Subject: ${clrbsub}"
        fi
    done
done


## ========================================================================================
## (check) We have the expected number of deployments
actual_deploys=$(oc get deployments | egrep "cert-manager-" | wc -l | sed -e 's/^[ \t]*//')
if [[ "${expected_deploys}" -ne "${actual_deploys}" ]]; then
    echo -e "${CRS} FAIL (Deployments): Expected ${expected_deploys}, Found ${actual_deploys}"
else
    echo -e "${CHK} PASS (Deployments)"
fi


## ========================================================================================
## (check) Lazy check that all deployments are ready
ready_deploys=$(oc get deployments --no-headers | egrep "cert-manager-" | grep '1/1' | wc -l | sed -e 's/^[ \t]*//')
if [[ "${expected_deploys}" -ne "${ready_deploys}" ]]; then
    echo -e "${CRS} FAIL (Deployments:AreReady): Expected ${expected_deploys}, Found ${ready_deploys}"
else
    echo -e "${CHK} PASS (Deployments:AreReady)"
fi