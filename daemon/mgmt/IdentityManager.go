// Package mgmt
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/26 8:52 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"minlib/common"
	"minlib/component"
	"minlib/mgmt"
	"minlib/minsecurity"
	cert2 "minlib/minsecurity/crypto/cert"
	"minlib/packet"
	"minlib/security"
	"strconv"
	"time"
)

// IdentityManager 表示一个身份管理器
// @Description:
//
type IdentityManager struct {
	keyChain *security.KeyChain
}

// CreateIdentityManager 创建一个 IdentityManager
//
// @Description:
// @param keyChain
// @return *IdentityManager
//
func CreateIdentityManager(keyChain *security.KeyChain) *IdentityManager {
	identityManager := new(IdentityManager)
	identityManager.keyChain = keyChain
	return identityManager
}

// Init 初始化
//
// @Description:
// @receiver im
// @param keyChain
//
func (im *IdentityManager) Init(dispatcher *Dispatcher) {

	// /identity-mgmt/add
	if identifier, err := component.CreateIdentifierByStringArray(mgmt.ManagementModuleIdentityMgmt,
		mgmt.IdentityManagementActionAdd); err != nil {
		common.LogFatal(err)
	} else {
		if err := dispatcher.AddControlCommand(identifier, dispatcher.authorization, func(parameters *component.ControlParameters) bool {
			return parameters.ControlParameterPrefix.IsInitial() &&
				parameters.ControlParameterPasswd.IsInitial()
		}, im.AddIdentity); err != nil {
			common.LogFatal(err)
		}
	}

	// /identity-mgmt/del
	if identifier, err := component.CreateIdentifierByStringArray(mgmt.ManagementModuleIdentityMgmt,
		mgmt.IdentityManagementActionDel); err != nil {
		common.LogFatal(err)
	} else {
		if err := dispatcher.AddControlCommand(identifier, dispatcher.authorization, func(parameters *component.ControlParameters) bool {
			return parameters.ControlParameterPrefix.IsInitial() &&
				parameters.ControlParameterPasswd.IsInitial()
		}, im.DelIdentity); err != nil {
			common.LogFatal(err)
		}
	}

	// /identity-mgmt/list
	if identifier, err := component.CreateIdentifierByStringArray(mgmt.ManagementModuleIdentityMgmt,
		mgmt.IdentityManagementActionList); err != nil {
		common.LogFatal(err)
	} else {
		if err := dispatcher.AddStatusDataset(
			identifier,
			dispatcher.authorization,
			func(parameters *component.ControlParameters) bool {
				return true
			},
			im.ListIdentity,
		); err != nil {
			common.LogFatal(err)
		}
	}

	// /identity-mgmt/dumpCert
	if identifier, err := component.CreateIdentifierByStringArray(mgmt.ManagementModuleIdentityMgmt,
		mgmt.IdentityManagementActionDumpCert); err != nil {
		common.LogFatal(err)
	} else {
		if err := dispatcher.AddStatusDataset(identifier, dispatcher.authorization,
			func(parameters *component.ControlParameters) bool {
				return parameters.ControlParameterPrefix.IsInitial()
			},
			im.DumpCert); err != nil {
			common.LogFatal(err)
		}
	}

	// /identity-mgmt/importCert
	if identifier, err := component.CreateIdentifierByStringArray(mgmt.ManagementModuleIdentityMgmt,
		mgmt.IdentityManagementActionImportCert); err != nil {
		common.LogFatal(err)
	} else {
		if err := dispatcher.AddControlCommand(identifier, dispatcher.authorization,
			func(parameters *component.ControlParameters) bool {
				return true
			}, im.ImportCert); err != nil {
			common.LogFatal(err)
		}
	}

	// /identity-mgmt/setDef
	if identifier, err := component.CreateIdentifierByStringArray(mgmt.ManagementModuleIdentityMgmt,
		mgmt.IdentityManagementActionSetDef); err != nil {
		common.LogFatal(err)
	} else {
		if err := dispatcher.AddControlCommand(identifier, dispatcher.authorization,
			func(parameters *component.ControlParameters) bool {
				return true
			}, im.SetDef); err != nil {
			common.LogFatal(err)
		}
	}

	// /identity-mgmt/dumpId
	if identifier, err := component.CreateIdentifierByStringArray(mgmt.ManagementModuleIdentityMgmt,
		mgmt.IdentityManagementActionDumpId); err != nil {
		common.LogFatal(err)
	} else {
		if err := dispatcher.AddStatusDataset(
			identifier,
			dispatcher.authorization,
			func(parameters *component.ControlParameters) bool {
				return parameters.ControlParameterPrefix.IsInitial()
			},
			im.DumpId,
		); err != nil {
			common.LogFatal(err)
		}
	}

	// /identity-mgmt/loadId
	if identifier, err := component.CreateIdentifierByStringArray(mgmt.ManagementModuleIdentityMgmt,
		mgmt.IdentityManagementActionLoadId); err != nil {
		common.LogFatal(err)
	} else {
		if err := dispatcher.AddControlCommand(identifier, dispatcher.authorization,
			func(parameters *component.ControlParameters) bool {
				return true
			}, im.LoadId); err != nil {
			common.LogFatal(err)
		}
	}

	// /identity-mgmt/getId
	if identifier, err := component.CreateIdentifierByStringArray(mgmt.ManagementModuleIdentityMgmt,
		mgmt.IdentityManagementActionGetId); err != nil {
		common.LogFatal(err)
	} else {
		if err := dispatcher.AddStatusDataset(
			identifier,
			dispatcher.authorization,
			func(parameters *component.ControlParameters) bool {
				return parameters.ControlParameterPrefix.IsInitial()
			},
			im.GetId,
		); err != nil {
			common.LogFatal(err)
		}
	}

	// /identity-mgmt/selfIssue
	if identifier, err := component.CreateIdentifierByStringArray(mgmt.ManagementModuleIdentityMgmt,
		mgmt.IdentityManagementActionSelfIssue); err != nil {
		common.LogFatal(err)
	} else {
		if err := dispatcher.AddControlCommand(identifier, dispatcher.authorization,
			func(parameters *component.ControlParameters) bool {
				return parameters.ControlParameterPrefix.IsInitial() &&
					parameters.ControlParameterPasswd.IsInitial()
			}, im.SelfIssue); err != nil {
			common.LogFatal(err)
		}
	}
}

// AddIdentity 添加一个网络身份
//
// @Description:
// @receiver im
// @param topPrefix
// @param interest
// @param parameters
// @return *mgmt.ControlResponse
//
func (im *IdentityManager) AddIdentity(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {
	if id, err := im.keyChain.CreateIdentityByName(parameters.Prefix().ToUri(), parameters.Passwd()); err != nil {
		return MakeControlResponse(mgmt.ControlResponseCodeCommonError, err.Error(), "")
	} else {
		return MakeControlResponse(mgmt.ControlResponseCodeSuccess, "Add identity success => "+id.Name, "")
	}
}

// DelIdentity 删除一个网络身份
//
// @Description:
// @receiver im
// @param topPrefix
// @param interest
// @param parameters
// @return *mgmt.ControlResponse
//
func (im *IdentityManager) DelIdentity(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {
	targetName := parameters.Prefix().ToUri()
	// 不允许删除当前正在使用的网络身份
	if currentIdentity := im.keyChain.GetCurrentIdentity(); currentIdentity != nil && currentIdentity.Name == targetName {
		return MakeControlResponse(mgmt.ControlResponseCodeCommonError, "Not allow delete current use identity!", "")
	}

	// 判断要删除的网络身份是否存在
	targetIdentity := im.keyChain.GetIdentityByName(targetName)
	if targetIdentity == nil {
		return MakeControlResponse(mgmt.ControlResponseCodeCommonError, "Identity not exists!", "")
	}

	// 判断密码是否正确
	if _, err := targetIdentity.UnLock(parameters.Passwd(), minsecurity.SM4ECB); err != nil {
		return MakeControlResponse(mgmt.ControlResponseCodeCommonError, err.Error(), "")
	}

	// 验证完成之后，再加锁回去
	if _, err := targetIdentity.Lock(parameters.Passwd(), minsecurity.SM4ECB); err != nil {
		return MakeControlResponse(mgmt.ControlResponseCodeCommonError, err.Error(), "")
	}

	if ok, err := im.keyChain.DeleteIdentityByName(targetName, parameters.Passwd()); err != nil {
		return MakeControlResponse(mgmt.ControlResponseCodeCommonError, err.Error(), "")
	} else {
		return MakeControlResponse(mgmt.ControlResponseCodeSuccess, "Delete identity => "+strconv.FormatBool(ok), "")
	}
}

// ListIdentityInfo 表示调用 ListIdentity 时，返回的item的数据结构
// @Description:
//
type ListIdentityInfo struct {
	Name string
}

// ListIdentity 列出所有的网络身份
//
// @Description:
// @receiver im
// @param topPrefix
// @param interest
// @param context
//
func (im *IdentityManager) ListIdentity(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters,
	context *StatusDatasetContext) {
	defaultName := ""
	if defaultIdentity := im.keyChain.GetDefaultIdentity(); defaultIdentity != nil {
		defaultName = defaultIdentity.Name
	}
	for _, v := range im.keyChain.GetAllIdentities() {
		if v.Name == defaultName {
			context.Append(&ListIdentityInfo{
				Name: "*" + v.Name,
			})
		} else {
			context.Append(&ListIdentityInfo{
				Name: v.Name,
			})
		}
	}
	_ = context.Done(im.keyChain.IdentityManager.GetCurrentVersion())
}

// DumpCert 导出某个指定身份的证书
//
// @Description:
// @receiver im
// @param topPrefix
// @param interest
// @param context
//
func (im *IdentityManager) DumpCert(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters,
	context *StatusDatasetContext) {
	common.LogDebug(">>>")
	identityName := parameters.Prefix().ToUri()

	// 首先判断指定的网络身份是否存在
	targetIdentity := im.keyChain.GetIdentityByName(identityName)
	if targetIdentity == nil {
		context.responseSender(
			MakeControlResponse(mgmt.ControlResponseCodeCommonError, "Target identity not exists!", ""),
			interest,
		)
		return
	}

	// 判断证书是否存在
	if targetIdentity.Cert.Issuer == "" && targetIdentity.Cert.Signature == nil {
		context.responseSender(
			MakeControlResponse(mgmt.ControlResponseCodeCommonError, "Target identity's cert not exists!", ""),
			interest,
		)
		return
	}

	// 导出证书
	if str, err := (&targetIdentity.Cert).ToPem([]byte(""), minsecurity.SM4ECB); err != nil {
		context.responseSender(
			MakeControlResponse(mgmt.ControlResponseCodeCommonError, err.Error(), ""),
			interest,
		)
		return
	} else {
		context.Append(str)
	}
	_ = context.Done(0)
}

// ImportCert 导入证书
//
// @Description:
// @receiver im
// @param topPrefix
// @param interest
// @param parameters
// @return *mgmt.ControlResponse
//
func (im *IdentityManager) ImportCert(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {
	return nil
}

// SetDef 设置默认的网络身份
//
// @Description:
// @receiver im
// @param topPrefix
// @param interest
// @param parameters
// @return *mgmt.ControlResponse
//
func (im *IdentityManager) SetDef(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {
	return nil
}

// DumpId 导出用户的身份
//
// @Description:
// @receiver im
// @param topPrefix
// @param interest
// @param context
//
func (im *IdentityManager) DumpId(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters,
	context *StatusDatasetContext) {

}

// LoadId 导入用户身份
//
// @Description:
// @receiver im
// @param topPrefix
// @param interest
// @param parameters
// @return *mgmt.ControlResponse
//
func (im *IdentityManager) LoadId(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {
	return nil
}

// GetId 获得一个指定网络身份的JSON序列化表示
//
// @Description:
// @receiver im
// @param topPrefix
// @param interest
// @param parameters
// @return *mgmt.ControlResponse
//
func (im *IdentityManager) GetId(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters,
	context *StatusDatasetContext) {
}

// SelfIssue 给当前网络身份
//
// @Description:
// @receiver im
// @param topPrefix
// @param interest
// @param parameters
// @return *mgmt.ControlResponse
//
func (im *IdentityManager) SelfIssue(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *component.ControlParameters) *mgmt.ControlResponse {
	identityName := parameters.Prefix().ToUri()
	passwd := parameters.Passwd()

	id := im.keyChain.GetIdentityByName(identityName)

	// 判断身份是否存在
	if id == nil {
		return MakeControlResponse(mgmt.ControlResponseCodeCommonError, "Target identity not exists!", "")
	}

	// 解锁身份
	if id.IsLocked() {
		if _, err := id.UnLock(passwd, minsecurity.SM4ECB); err != nil {
			return MakeControlResponse(mgmt.ControlResponseCodeCommonError, err.Error(), "")
		}
	}

	// 填充证书内容
	cert := cert2.Certificate{}
	cert.Version = 0
	cert.SerialNumber = 1
	cert.PublicKey = id.Pubkey
	cert.SignatureAlgorithm = id.KeyParam.SignatureAlgorithm
	cert.PublicKeyAlgorithm = id.KeyParam.PublicKeyAlgorithm
	cert.IssueTo = id.Name
	cert.Issuer = id.Name
	cert.NotBefore = time.Now().Unix()
	cert.NotAfter = time.Now().AddDate(1, 0, 0).Unix()
	cert.KeyUsage = minsecurity.CertSign
	cert.IsCA = true
	cert.Timestamp = time.Now().Unix()

	// 签发证书
	if err := cert.SignCert(id.Prikey); err != nil {
		return MakeControlResponse(mgmt.ControlResponseCodeCommonError, err.Error(), "")
	}

	id.Cert = cert

	// 持久化保存
	if err := im.keyChain.SaveIdentity(id, true); err != nil {
		return MakeControlResponse(mgmt.ControlResponseCodeCommonError, err.Error(), "")
	}

	return MakeControlResponse(mgmt.ControlResponseCodeSuccess, "", "")
}
