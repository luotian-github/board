import os
import sys
import commands
import jenkins
import json
import time
import shutil
import requests

import libvirtapi

kmvNameList = ['kvm-1', 'kvm-2', 'kvm-3', 'kvm-4']
tmpkvmdir = '/tmp/kvm'

basedir = os.path.dirname(os.path.realpath(__file__))

credentialId = 'k-v-m-i-d'

kvmnodeUserName = 'root'
kvmnodePasswd = '123456a?'

def createCredential(jenkinsMaster, nodeUsername, nodePasswd, credentialId):
    para = "--data-urlencode"
    post_url = '%s/credentials/store/system/domain/_/createCredentials ' %jenkinsMaster
    credent = {
    "scope": "GLOBAL",
    "id": credentialId,
    "username": nodeUsername,
    "password": nodePasswd,
    "description": "auto added",
    "$class": "com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl"}

    data = json.dumps(credent)
    cmd = '''curl -X POST %s %s 'json={"":"0","credentials":%s}'
          ''' %(post_url, para, data)
    print (cmd)
    os.system(cmd)


def getNodeIp(nodename):
    macline = commands.getoutput('virsh dumpxml %s|grep mac|grep address' %nodename)
    mac = macline.split("'")[1]
    ipaddr = commands.getoutput('''arp -a|grep %s|cut -d '(' -f2|cut -d ')' -f1''' %mac)
    return ipaddr
def checkNode(jenkinsServer, usekvmnode):
    try:
        jenkinsServer.assert_node_exists(usekvmnode, exception_message='node %s does not exist.......')
    except jenkins.JenkinsException, e:
        print (e)
    nodeflag = jenkinsServer.node_exists(usekvmnode) 
    return nodeflag

def addJenkinsNode(jenkinsMaster, nodename):
    ipaddr = getNodeIp(nodename)
    while len(ipaddr)<7:
        time.sleep(0.5)
        #ipaddr = getNodeIp()
        _, ipaddr=commands.getstatusoutput("%s/search.sh %s" %(basedir,nodename))
    print ("==============:::" + ipaddr)
    cid = credentialId
    server = jenkins.Jenkins(jenkinsMaster)
    flag = checkNode(server, nodename)
    
    while flag == True:
        time.sleep(2)
        flag = checkNode(server, nodename)
    
    params = {
        'port': 22,
        'username': 'juser',
        'credentialsId': cid,
        'host': ipaddr
    }
    print ("params: %s" %params)
    try:
        server.create_node(
            nodename,nodeDescription='add slave',
            remoteFS='/data',
            numExecutors=3,
            labels='slave',
            exclusive=False,
            launcher=jenkins.LAUNCHER_SSH,
            launcher_params=params)
    except Exception, e:
       print 'str(e):\t\t', str(e)
       print ("failed add jenkins node")
def checkKVM(flagFile):
    if os.path.isfile(flagFile):
        with open(flagFile, 'r') as f:
            lines = f.readlines()
            flag = lines[0].replace("\n", "")
            f.close
    print ("==%s==" %flag)
    return flag

def copyImage(kvmName):
    baseimage='/var/lib/libvirt/images/kvm.img'
    image = '/var/lib/libvirt/images/%s.img' %kvmName
    if os.path.exists(image):
        os.remove(image)
    else:
        print ('no image can be cleaned!!!')
    if os.path.exists(baseimage):
        shutil.copyfile(baseimage,image)
    else:
        print ('no base image can be copied!!!')

def diff(listA,listB):
    notused = list(set(listB).difference(set(listA)))
    return notused

def getdifflist():
    listB = kmvNameList
    listA = getlistofkvmDir()
    notusedkvmname = diff(listA,listB)
    return notusedkvmname

def getNumberofkvm():
    return len(os.listdir(tmpkvmdir))
        
def getlistofkvmDir():
    for _, _, files in os.walk(tmpkvmdir):
        usingKvmList = files
    return usingKvmList

def getKvmName(kvmApiServer, projectName):
    url = '%s/register-job' %kvmApiServer
    data = {'job_name': projectName}
    res = requests.post(url, data=data)
    kvmName = res.text
    return kvmName

#curl "http://10.165.15.148:8888/job/$1/buildWithParameters?node_name=$2"
def triggerKvmJob(jenkinsMaster, jobName, kvmName):
    url = '%s/job/%s/buildWithParameters' % (jenkinsMaster, jobName)
    data = {'node_name': kvmName}
    res = requests.get(url, params=data)
    return res.text

def startKVM(jenkinsMaster, kvmName, jobName, kvmApiServer):
    # kvmName = getKvmName(kvmApiServer, projectName)    
    print ("::::" + kvmName + ":::")
    # while kvmName == 'FULL':
    #     time.sleep(3)
    #     print ('kvm is full, can not get a kvm name...')
    #     kvmName = getKvmName(kvmApiServer, jobName)

    usekvmname = kvmName
    print(usekvmname)

    conn = libvirtapi.createConnection()
    myDom = libvirtapi.getDomInfoByName(conn, usekvmname)
    libvirtapi.closeConnection(conn)
    
    if myDom == 1:
        print ('Create KVM ...........................')
        copyImage(usekvmname)
        try:
            os.popen("virt-install --name %s --ram 2048 --disk path=/var/lib/libvirt/images/%s.img --import &\n\n\n" %(usekvmname, usekvmname))
        except:
            print('Failed to create KVM')
    else:
        print ('Revert snapshot .........................')
        try:
            os.popen("virsh snapshot-revert %s %s" %(usekvmname, usekvmname))
        except:
            print('Failed to create KVM')
    addJenkinsNode(jenkinsMaster,usekvmname)
    triggerKvmJob(jenkinsMaster, jobName, usekvmname)
    
    if myDom == 1:
        try:
            os.popen('virsh snapshot-create-as %s %s' %(usekvmname, usekvmname))
        except:
            print('Failed to create KVM')
    

def startService():
    os.system('systemctl start docker')

def main():
    jenkinsMaster = sys.argv[1]
    projectName = sys.argv[2]
    nodeName = sys.argv[3]
    kvmApiServer = sys.argv[4]
    createCredential(jenkinsMaster, kvmnodeUserName, kvmnodePasswd, credentialId)
    startKVM(jenkinsMaster, nodeName, projectName, kvmApiServer)
    print ("add jenkins node ..................")
   

if __name__ == "__main__":
    main()
