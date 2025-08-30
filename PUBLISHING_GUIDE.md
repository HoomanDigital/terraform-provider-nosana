# Terraform Registry Publishing Guide

This guide outlines the steps to publish your Nosana Terraform Provider to the official Terraform Registry.

## ✅ Completed Steps

The following files have been created and configured for registry submission:

1. **LICENSE** - MIT license required by registry
2. **Acceptance Tests** - Located in `nosana/` directory
3. **GitHub Actions** - CI/CD workflows for testing and releases
4. **GoReleaser** - Automated release configuration
5. **Documentation** - Registry-compliant docs in `docs/` directory
6. **Module Path** - Updated to match GitHub repository

## 🚀 Next Steps for Registry Publication

### Step 1: Repository Setup

First, ensure your GitHub repository is public and push all the new files:

```bash
# Add all new files
git add .
git commit -m "Prepare provider for Terraform Registry publication"
git push origin main
```

### Step 2: GPG Key Setup (Required for Signing)

The Terraform Registry requires signed releases. Set up GPG signing:

1. **Generate a GPG key** (if you don't have one):
   ```bash
   gpg --full-generate-key
   # Choose RSA, 4096 bits, set expiration, provide name/email
   ```

2. **Export your public key**:
   ```bash
   gpg --armor --export your-email@example.com > public.key
   ```

3. **Add GPG secrets to GitHub**:
   - Go to your repository Settings → Secrets and variables → Actions
   - Add these secrets:
     - `GPG_PRIVATE_KEY`: Your private key (`gpg --armor --export-secret-keys your-email@example.com`)
     - `PASSPHRASE`: Your GPG key passphrase

### Step 3: Create Your First Release

1. **Tag and push a release**:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

2. **Verify the release**:
   - Go to your GitHub repository
   - Check the "Actions" tab to ensure the release workflow runs successfully
   - Verify that a release with signed binaries appears in "Releases"

### Step 4: Submit to Terraform Registry

1. **Go to the Terraform Registry**:
   - Visit [registry.terraform.io](https://registry.terraform.io)
   - Click "Publish" → "Provider"

2. **Connect your GitHub account** and select your repository:
   - Repository: `HoomanDigital/TerraformProvider-Nosana`

3. **Registry Requirements Check**:
   The registry will automatically verify:
   - ✅ Repository is public
   - ✅ Has a valid license
   - ✅ Has signed releases
   - ✅ Has proper documentation structure
   - ✅ Follows naming convention (`terraform-provider-*`)

4. **Repository Naming** (⚠️ Important):
   Your repository needs to be renamed to follow Terraform conventions:
   ```
   Current: TerraformProvider-Nosana
   Required: terraform-provider-nosana
   ```

   **Action needed**: Rename your GitHub repository to `terraform-provider-nosana`

### Step 5: Update Configuration After Rename

After renaming the repository, update these files:

1. **Update go.mod**:
   ```go
   module github.com/HoomanDigital/terraform-provider-nosana
   ```

2. **Update main.go import**:
   ```go
   "github.com/HoomanDigital/terraform-provider-nosana/nosana"
   ```

3. **Update .goreleaser.yml** (if needed for project references)

### Step 6: Testing Before Submission

Test your provider thoroughly:

```bash
# Run tests
go test ./nosana/

# Test with real Terraform
terraform init
terraform plan
terraform apply
```

### Step 7: Registry Verification Process

Once submitted, the Terraform Registry will:

1. **Verify repository structure** (automated)
2. **Check for required files** (automated)
3. **Validate releases and signatures** (automated)
4. **Review provider functionality** (manual, may take 1-3 business days)

## 📋 Registry Requirements Checklist

- ✅ Public GitHub repository
- ⚠️ Repository named `terraform-provider-nosana` (needs rename)
- ✅ Valid license file
- ✅ Signed releases with GoReleaser
- ✅ Acceptance tests
- ✅ Documentation in `docs/` directory
- ✅ GitHub Actions for CI/CD
- ✅ Go module with correct path

## 🔧 Post-Publication Steps

After your provider is approved:

1. **Update documentation** to reference the registry source:
   ```hcl
   terraform {
     required_providers {
       nosana = {
         source = "HoomanDigital/nosana"
       }
     }
   }
   ```

2. **Update README.md** with registry installation instructions

3. **Create release notes** for future versions

4. **Set up monitoring** for provider usage and feedback

## 📚 Additional Resources

- [Terraform Registry Provider Publishing](https://developer.hashicorp.com/terraform/registry/providers/publishing)
- [Provider Development Documentation](https://developer.hashicorp.com/terraform/plugin/sdkv2)
- [GoReleaser Documentation](https://goreleaser.com/)

## 🆘 Troubleshooting

### Common Issues:

1. **GPG Signing Fails**:
   - Ensure GPG_PRIVATE_KEY secret is properly formatted
   - Check that PASSPHRASE is correct

2. **Repository Not Found**:
   - Verify repository name follows `terraform-provider-*` convention
   - Ensure repository is public

3. **Release Workflow Fails**:
   - Check GitHub Actions logs
   - Verify all required secrets are set

4. **Registry Rejection**:
   - Review feedback from HashiCorp team
   - Ensure all requirements are met
   - Check that tests pass

## 📞 Support

If you encounter issues:
- Check the [Terraform Provider Registry Support](https://developer.hashicorp.com/terraform/registry/providers/publishing#support)
- Review provider requirements carefully
- Test thoroughly before submission
