package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// commands is a register of all availables commands that can be executed by
// this program. The name is used to match with the first argument given.
//
// When a cmd function is callend it is given stdin, stdout and command line
// arguments except the program name and this command name. It is the
// responsibility of the command function to parse the arguments. Use os.Stderr
// to write error messages.
//
// A command function is an independent runable that is taking input and output
// being stdin and stdout. Given args are the command line arguments, without
// the program name, that should be parsed using the flag package.
// A command function is expected to read and write only to provided input and
// output. In a special case of an invalid argument a message to os.Stderr and
// os.Exit(2) call are allowed.
//
// When implementing a command function, keep it simple. A command function
// should provide a single functionality. A unix pipe can be used to construct
// a pipeline. For example, there are 3 separate functions for creating a
// transaction, signing and submitting. They can be combined into a single
// pipeline line:
//
//   $ bnscli release-escrow -escrow 1 \
//       | bnscli as-proposal \
//       | bnscli sign \
//       | bnscli submit
//
var commands = map[string]func(input io.Reader, output io.Writer, args []string) error{
	// This cannot be registered here because of circular reference.
	// "zsh-completion":         cmdZshCompletion,

	"add-account-certificate":              cmdAddAccountCertificate,
	"as-batch":                             cmdAsBatch,
	"as-proposal":                          cmdAsProposal,
	"as-sequence":                          cmdAsSequence,
	"datamigration":                        cmdDataMigrationExecute,
	"del-account-certificate":              cmdDelAccountCertificate,
	"del-proposal":                         cmdDelProposal,
	"delete-account":                       cmdDeleteAccount,
	"delete-domain":                        cmdDeleteDomain,
	"flush-domain":                         cmdFlushDomain,
	"from-sequence":                        cmdFromSequence,
	"keyaddr":                              cmdKeyaddr,
	"keygen":                               cmdKeygen,
	"mnemonic":                             cmdMnemonic,
	"msgfee-update-configuration":          cmdMsgFeeUpdateConfiguration,
	"multisig":                             cmdMultisig,
	"preregistration-register":             cmdPreregistrationRegister,
	"preregistration-update-configuration": cmdPreregistrationUpdateConfiguration,
	"qualityscore-update-configuration":    cmdQualityscoreUpdateConfiguration,
	"query":                                cmdQuery,
	"register-account":                     cmdRegisterAccount,
	"register-domain":                      cmdRegisterDomain,
	"register-username":                    cmdRegisterUsername,
	"release-escrow":                       cmdReleaseEscrow,
	"renew-account":                        cmdRenewAccount,
	"renew-domain":                         cmdRenewDomain,
	"replace-account-msg-fees":             cmdReplaceAccountMsgFees,
	"replace-account-targets":              cmdReplaceAccountTrarget,
	"reset-revenue":                        cmdResetRevenue,
	"resolve-username":                     cmdResolveUsername,
	"send-tokens":                          cmdSendTokens,
	"set-msgfee":                           cmdSetMsgFee,
	"set-validators":                       cmdSetValidators,
	"sign":                                 cmdSignTransaction,
	"submit":                               cmdSubmitTransaction,
	"termdeposit-create-contract":          cmdTermdepositCreateDepositContract,
	"termdeposit-deposit":                  cmdTermdepositDeposit,
	"termdeposit-release-deposit":          cmdTermdepositReleaseDeposit,
	"termdeposit-update-configuration":     cmdTermdepositUpdateConfiguration,
	"termdeposit-with-base-rate":           cmdTermdepositWithBaseRate,
	"termdeposit-with-bonus":               cmdTermdepositWithBonus,
	"text-resolution":                      cmdTextResolution,
	"transfer-account":                     cmdTransferAccount,
	"transfer-domain":                      cmdTransferDomain,
	"txfee-print-rates":                    cmdTxfeePrintRates,
	"txfee-update-configuration":           cmdTxfeeUpdateConfiguration,
	"update-account-configuration":         cmdUpdateAccountConfiguration,
	"update-cash-configuration":            cmdUpdateCashConfiguration,
	"update-election-rule":                 cmdUpdateElectionRule,
	"update-electorate":                    cmdUpdateElectorate,
	"update-username-configuration":        cmdUpdateUsernameConfiguration,
	"upgrade-schema":                       cmdUpgradeSchema,
	"version":                              cmdVersion,
	"view":                                 cmdTransactionView,
	"vote":                                 cmdVote,
	"with-account-msg-fee":                 cmdWithAccountMsgFee,
	"with-account-target":                  cmdWithAccountTarget,
	"with-blockchain-address":              cmdWithBlockchainAddress,
	"with-elector":                         cmdWithElector,
	"with-fee":                             cmdWithFee,
	"with-multisig":                        cmdWithMultisig,
	"with-multisig-participant":            cmdWithMultisigParticipant,
}

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "%s is a command line client for the BNSD application.\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [<flags>]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nAvailable commands are:\n\t%s\n", strings.Join(availableCmds(), "\n\t"))
		fmt.Fprintf(os.Stderr, "Run '%s <command> -help' to learn more about each command.\n", os.Args[0])
		os.Exit(2)
	}
	run, ok := commands[os.Args[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", os.Args[1])
		fmt.Fprintf(os.Stderr, "\nAvailable commands are:\n\t%s\n", strings.Join(availableCmds(), "\n\t"))
		os.Exit(2)
	}

	// Skip two first arguments. Second argument is the command name that
	// we just consumed.
	if err := run(os.Stdin, os.Stdout, os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func availableCmds() []string {
	available := make([]string, 0, len(commands))
	for name := range commands {
		available = append(available, name)
	}
	sort.Strings(available)
	return available
}

func cmdVersion(in io.Reader, out io.Writer, args []string) error {
	fmt.Fprintln(out, gitHash)
	return nil
}

// gitHash is set during the compilation time.
var gitHash string = "dev"
