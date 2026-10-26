package profile

// profileWarnings lists settings that are valid but risky or easy to
// misunderstand. They never block generation.
func profileWarnings(p *Profile) []string {
	var w []string

	hasPassword := p.FirstLogon.BuiltinAdministratorPassword != nil
	for _, a := range p.Accounts {
		if a.Password != nil {
			hasPassword = true
		}
	}
	if hasPassword && !p.ObscurePasswords {
		w = append(w, "Пароли попадут в autounattend.xml открытым текстом. Не храните файл на общедоступных носителях и в репозитории.")
	}
	if hasPassword && p.ObscurePasswords {
		w = append(w, "obscure_passwords лишь кодирует пароли в Base64, это не шифрование: любой, у кого есть файл, прочитает пароли.")
	}
	if hasPassword && p.KeepSensitiveFiles {
		w = append(w, "keep_sensitive_files оставит на установленной системе копию ответного файла с паролями (C:\\Windows\\Panther).")
	}
	if p.FirstLogon.Mode == FirstLogonBuiltinAdmin {
		w = append(w, "Встроенная учётная запись Administrator будет включена и использована для первого входа.")
	}
	if p.SystemTweaks.DisableUAC {
		w = append(w, "disable_uac отключает контроль учётных записей (UAC): это снижает безопасность системы.")
	}
	if p.SystemTweaks.DisableWindowsUpdate {
		w = append(w, "disable_windows_update отключает автоматические обновления: система не будет получать исправления безопасности.")
	}
	if p.SystemTweaks.HardenSystemDriveACL {
		w = append(w, "harden_system_drive_acl меняет права доступа к системному диску и может помешать работе некоторых программ.")
	}
	if p.SystemTweaks.BypassWin11Requirements {
		w = append(w, "bypass_win11_requirements обходит проверки оборудования Windows 11: такая установка не поддерживается Microsoft.")
	}
	if p.BypassOnlineAccountRequirement {
		w = append(w, "bypass_online_account_requirement работает не на всех сборках Windows: Microsoft неоднократно закрывала этот обход.")
	}
	return w
}
